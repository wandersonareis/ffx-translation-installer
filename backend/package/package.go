// Package ffxpackage implements FF13 package format v2 extraction for FFXX2HDR.
//
// This package only supports extraction (reading) of v2 packages with the
// magic identifier "FFXX2HDR". It does not support packing, v1 format, or
// any other magic values.
package ffxpackage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/cespare/xxhash"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
)

// MagicFFXX2HDR is the package magic identifier for FINAL FANTASY X/X-2 HD Remaster translation packages.
const MagicFFXX2HDR = "FFXX2HDR"

const (
	// VersionV2 is the format version number.
	VersionV2 = 2
	// HeaderSizeV2 is the fixed header size (64 bytes).
	HeaderSizeV2 = 64
	// MaxFileNameLen is the sanity limit for stored names.
	MaxFileNameLen = 1024
	// EntryFixedLen is the fixed per-entry size in the file table (48 bytes).
	EntryFixedLen = 48
)

// FlagUsesDict marks entries whose zstd payload was compressed with the
// package dictionary. The reader must apply the dictionary only for these.
const FlagUsesDict uint16 = 1 << 0

type CompressionAlgorithm uint8

const (
	CompNone CompressionAlgorithm = iota
	CompZstd
	CompLZ4
	CompBrotli
	// CompAuto is a packer-only pseudo algorithm: never written into a package.
	CompAuto CompressionAlgorithm = 255
)

func (c CompressionAlgorithm) String() string {
	switch c {
	case CompNone:
		return "none"
	case CompZstd:
		return "zstd"
	case CompLZ4:
		return "lz4"
	case CompBrotli:
		return "brotli"
	case CompAuto:
		return "auto"
	default:
		return "unknown"
	}
}

var (
	ErrInvalidMagicV2   = errors.New("invalid v2 magic number")
	ErrInvalidVersionV2 = errors.New("invalid v2 version")
	ErrUnsupportedComp  = errors.New("unsupported compression algorithm")
	ErrCorruptedDataV2  = errors.New("corrupted v2 data")
)

// HeaderV2 is the 64-byte package header.
type HeaderV2 struct {
	Magic           [8]byte
	Version         uint32
	Flags           uint32
	FileCount       uint32
	FileTableOffset uint64
	DictOffset      uint64
	DictSize        uint32 // as stored on disk
	DictStorage     CompressionAlgorithm
	Reserved        [23]byte
}

func (h *HeaderV2) IsValid() bool {
	return string(h.Magic[:]) == MagicFFXX2HDR && h.Version == VersionV2
}

// FileEntryV2 describes one stored file.
type FileEntryV2 struct {
	Name             string
	UncompressedSize uint64
	CompressedSize   uint64
	Offset           uint64
	Algorithm        CompressionAlgorithm
	CompressionLevel int
	Checksum         uint32 // CRC32 (IEEE) of uncompressed data
	Hash             [8]byte // xxHash64 of uncompressed data
	Flags            uint16
	ModTime          int64 // unix seconds
}

// UsesDict reports whether this entry needs the package zstd dictionary.
func (e *FileEntryV2) UsesDict() bool { return e.Flags&FlagUsesDict != 0 }

// PackageV2 is an opened v2 package (read-only, extraction only).
type PackageV2 struct {
	Header HeaderV2
	Files  []FileEntryV2
	Dict   []byte // decoded dictionary content (nil when absent)
	reader io.ReaderAt
	size   int64
}

// Open opens and parses a v2 package with the FFXX2HDR magic.
// This is the only supported magic for this package.
func Open(filename string) (*PackageV2, error) {
	return OpenWithMagic(filename, MagicFFXX2HDR)
}

// OpenWithMagic opens and parses a v2 package expecting a custom magic.
// For this package, only MagicFFXX2HDR is valid.
func OpenWithMagic(filename, magic string) (*PackageV2, error) {
	if magic != MagicFFXX2HDR {
		return nil, fmt.Errorf("unsupported magic %q: only %q is supported", magic, MagicFFXX2HDR)
	}
	if err := Validate(magic); err != nil {
		return nil, err
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	p := &PackageV2{reader: f, size: stat.Size()}
	if err := p.parseHeader(); err != nil {
		f.Close()
		return nil, err
	}
	if err := p.parseFileTable(); err != nil {
		f.Close()
		return nil, err
	}
	if p.Header.DictSize > 0 {
		if err := p.loadDictionary(); err != nil {
			f.Close()
			return nil, err
		}
	}
	return p, nil
}

// Close closes the underlying file.
func (p *PackageV2) Close() error {
	if f, ok := p.reader.(*os.File); ok {
		p.reader = nil
		return f.Close()
	}
	p.reader = nil
	return nil
}

// Size returns the package file size in bytes.
func (p *PackageV2) Size() int64 { return p.size }

func (p *PackageV2) parseHeader() error {
	buf := make([]byte, HeaderSizeV2)
	if _, err := p.reader.ReadAt(buf, 0); err != nil {
		return err
	}
	copy(p.Header.Magic[:], buf[0:8])
	p.Header.Version = binary.LittleEndian.Uint32(buf[8:12])
	p.Header.Flags = binary.LittleEndian.Uint32(buf[12:16])
	p.Header.FileCount = binary.LittleEndian.Uint32(buf[16:20])
	p.Header.FileTableOffset = binary.LittleEndian.Uint64(buf[20:28])
	p.Header.DictOffset = binary.LittleEndian.Uint64(buf[28:36])
	p.Header.DictSize = binary.LittleEndian.Uint32(buf[36:40])
	p.Header.DictStorage = CompressionAlgorithm(buf[40])
	copy(p.Header.Reserved[:], buf[41:64])

	if string(p.Header.Magic[:]) != MagicFFXX2HDR {
		return fmt.Errorf("invalid magic %q: expected %q (package not compatible with this reader)",
			string(p.Header.Magic[:]), MagicFFXX2HDR)
	}
	if p.Header.Version != VersionV2 {
		return fmt.Errorf("%w: %d", ErrInvalidVersionV2, p.Header.Version)
	}
	return nil
}

func (p *PackageV2) parseFileTable() error {
	if p.Header.FileTableOffset == 0 || p.Header.FileTableOffset >= uint64(p.size) {
		return fmt.Errorf("%w: file table offset %d", ErrCorruptedDataV2, p.Header.FileTableOffset)
	}
	if uint64(p.Header.FileCount) > 1_000_000 {
		return fmt.Errorf("%w: file count %d", ErrCorruptedDataV2, p.Header.FileCount)
	}

	p.Files = make([]FileEntryV2, 0, p.Header.FileCount)
	offset := int64(p.Header.FileTableOffset)

	for i := uint32(0); i < p.Header.FileCount; i++ {
		var lenBuf [2]byte
		if _, err := p.reader.ReadAt(lenBuf[:], offset); err != nil {
			return err
		}
		nameLen := binary.LittleEndian.Uint16(lenBuf[:])
		if nameLen == 0 || nameLen > MaxFileNameLen {
			return fmt.Errorf("%w: name length %d", ErrCorruptedDataV2, nameLen)
		}
		offset += 2

		nameBuf := make([]byte, nameLen)
		if _, err := p.reader.ReadAt(nameBuf, offset); err != nil {
			return err
		}
		name := string(nameBuf)
		offset += int64(nameLen)

		var entryBuf [EntryFixedLen]byte
		if _, err := p.reader.ReadAt(entryBuf[:], offset); err != nil {
			return err
		}
		offset += EntryFixedLen

		entry := FileEntryV2{
			Name:             name,
			UncompressedSize: binary.LittleEndian.Uint64(entryBuf[0:8]),
			CompressedSize:   binary.LittleEndian.Uint64(entryBuf[8:16]),
			Offset:           binary.LittleEndian.Uint64(entryBuf[16:24]),
			Algorithm:        CompressionAlgorithm(entryBuf[24]),
			CompressionLevel: int(entryBuf[25]),
			Checksum:         binary.LittleEndian.Uint32(entryBuf[26:30]),
			Flags:            binary.LittleEndian.Uint16(entryBuf[30:32]),
			ModTime:          int64(binary.LittleEndian.Uint64(entryBuf[32:40])),
		}
		copy(entry.Hash[:], entryBuf[40:48])
		p.Files = append(p.Files, entry)
	}
	return nil
}

func (p *PackageV2) loadDictionary() error {
	if p.Header.DictOffset == 0 || uint64(p.Header.DictOffset) >= uint64(p.size) {
		return fmt.Errorf("%w: bad dictionary offset", ErrCorruptedDataV2)
	}
	if uint64(p.Header.DictOffset)+uint64(p.Header.DictSize) > uint64(p.size) {
		return fmt.Errorf("%w: dictionary extends past end of file", ErrCorruptedDataV2)
	}

	blob := make([]byte, p.Header.DictSize)
	if _, err := p.reader.ReadAt(blob, int64(p.Header.DictOffset)); err != nil {
		return err
	}

	switch p.Header.DictStorage {
	case CompNone:
		p.Dict = blob
	case CompZstd:
		dec, err := zstd.NewReader(bytes.NewReader(blob))
		if err != nil {
			return err
		}
		defer dec.Close()
		p.Dict, err = io.ReadAll(dec)
		if err != nil {
			return fmt.Errorf("%w: dictionary decompress: %v", ErrCorruptedDataV2, err)
		}
	default:
		return fmt.Errorf("%w: dictionary storage %d", ErrCorruptedDataV2, p.Header.DictStorage)
	}

	// Validate the dictionary parses (must be a full zstd dictionary).
	if len(p.Dict) <= 8 || string(p.Dict[:4]) != "\x37\xa4\x30\xec" {
		return fmt.Errorf("%w: dictionary is not a valid zstd dictionary", ErrCorruptedDataV2)
	}
	return nil
}

// ReadFile returns the verified, decompressed content of one entry.
func (p *PackageV2) ReadFile(entry FileEntryV2) ([]byte, error) {
	if entry.CompressedSize == 0 {
		if entry.UncompressedSize != 0 {
			return nil, fmt.Errorf("%w: %s has no payload but size %d", ErrCorruptedDataV2, entry.Name, entry.UncompressedSize)
		}
		return nil, nil
	}

	compressed := make([]byte, entry.CompressedSize)
	if _, err := p.reader.ReadAt(compressed, int64(entry.Offset)); err != nil {
		return nil, fmt.Errorf("read %s: %w", entry.Name, err)
	}

	var data []byte
	var err error
	switch entry.Algorithm {
	case CompNone:
		data = compressed
	case CompZstd:
		data, err = p.decompressZstd(compressed, entry.UsesDict())
	case CompLZ4:
		data, err = decompressLZ4(compressed)
	case CompBrotli:
		data, err = decompressBrotli(compressed)
	default:
		return nil, fmt.Errorf("%w: %s (algorithm %d)", ErrUnsupportedComp, entry.Name, entry.Algorithm)
	}
	if err != nil {
		return nil, fmt.Errorf("decompress %s: %w", entry.Name, err)
	}

	if uint64(len(data)) != entry.UncompressedSize {
		return nil, fmt.Errorf("%w: %s size mismatch (expected %d, got %d)", ErrCorruptedDataV2, entry.Name, entry.UncompressedSize, len(data))
	}
	if crc32.ChecksumIEEE(data) != entry.Checksum {
		return nil, fmt.Errorf("%w: %s CRC32 mismatch", ErrCorruptedDataV2, entry.Name)
	}
	if xxhash.Sum64(data) != binary.LittleEndian.Uint64(entry.Hash[:]) {
		return nil, fmt.Errorf("%w: %s xxHash64 mismatch", ErrCorruptedDataV2, entry.Name)
	}
	return data, nil
}

// Extract unpacks every file into destDir, verifying integrity and restoring
// modification times.
func (p *PackageV2) Extract(destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for i := range p.Files {
		entry := &p.Files[i]
		data, err := p.ReadFile(*entry)
		if err != nil {
			return err
		}

		outPath := filepath.Join(destDir, entry.Name)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return err
		}
		if entry.ModTime != 0 {
			t := time.Unix(entry.ModTime, 0)
			if err := os.Chtimes(outPath, t, t); err != nil {
				return err
			}
		}
	}
	return nil
}

// List returns the parsed entries.
func (p *PackageV2) List() []FileEntryV2 { return p.Files }

func (p *PackageV2) decompressZstd(compressed []byte, usesDict bool) ([]byte, error) {
	var opts []zstd.DOption
	if usesDict {
		if len(p.Dict) == 0 {
			return nil, fmt.Errorf("%w: entry needs dictionary but package has none", ErrCorruptedDataV2)
		}
		opts = append(opts, zstd.WithDecoderDicts(p.Dict))
	}
	dec, err := zstd.NewReader(bytes.NewReader(compressed), opts...)
	if err != nil {
		return nil, err
	}
	defer dec.Close()
	return io.ReadAll(dec)
}

func decompressLZ4(compressed []byte) ([]byte, error) {
	r := lz4.NewReader(bytes.NewReader(compressed))
	return io.ReadAll(r)
}

func decompressBrotli(compressed []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(compressed))
	return io.ReadAll(r)
}