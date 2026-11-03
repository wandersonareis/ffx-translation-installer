package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	ffxpackage "ffxinstaller/backend/package"
)

// SourceEntry é uma entrada do pacote vista pelo instalador.
type SourceEntry struct {
	Name    string
	Size    int64
	ModTime time.Time
}

// Source abstrai o pacote para o instalador. Os testes usam uma fonte em
// memória; a produção usa *PackageSource sobre ffxpackage.PackageV2.
type Source interface {
	Info() PackageInfo
	Entries() []SourceEntry
	ReadFile(name string) ([]byte, error)
}

// PackageSource adapta um pacote v2 (FFXX2HDR) à interface Source.
type PackageSource struct {
	pkg    *ffxpackage.PackageV2
	info   PackageInfo
	byName map[string]int
}

// NewPackageSource valida o arquivo do pacote e monta os metadados que vão
// para o registro (inclusive o SHA-256 e o modtime do arquivo).
func NewPackageSource(pkg *ffxpackage.PackageV2, path string) (*PackageSource, error) {
	if pkg == nil {
		return nil, fmt.Errorf("pacote vazio")
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("lendo pacote %s: %w", path, err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("o pacote não é um arquivo: %s", path)
	}

	sum, err := hashFileSHA256(path)
	if err != nil {
		return nil, err
	}

	// Data da tradução = modtime do pacote. Sem um valor válido (alguns
	// sistemas de arquivos devolvem 0), usamos a entrada mais recente.
	modifiedAt := stat.ModTime()
	if modifiedAt.IsZero() || modifiedAt.Unix() <= 0 {
		modifiedAt = latestEntryModTime(pkg.Files)
	}

	byName := make(map[string]int, len(pkg.Files))
	for i := range pkg.Files {
		byName[pkg.Files[i].Name] = i
	}

	return &PackageSource{
		pkg: pkg,
		info: PackageInfo{
			Path:          path,
			Name:          filepath.Base(path),
			Size:          stat.Size(),
			SHA256:        sum,
			Magic:         string(pkg.Header.Magic[:]),
			FormatVersion: pkg.Header.Version,
			FileCount:     pkg.Header.FileCount,
			ModifiedAt:    modifiedAt,
		},
		byName: byName,
	}, nil
}

// Info devolve os metadados do pacote.
func (s *PackageSource) Info() PackageInfo { return s.info }

// Entries lista as entradas do pacote.
func (s *PackageSource) Entries() []SourceEntry {
	out := make([]SourceEntry, 0, len(s.pkg.Files))
	for i := range s.pkg.Files {
		entry := &s.pkg.Files[i]
		out = append(out, SourceEntry{
			Name:    entry.Name,
			Size:    int64(entry.UncompressedSize),
			ModTime: time.Unix(entry.ModTime, 0),
		})
	}
	return out
}

// ReadFile devolve o conteúdo verificado de uma entrada.
func (s *PackageSource) ReadFile(name string) ([]byte, error) {
	idx, ok := s.byName[name]
	if !ok {
		return nil, fmt.Errorf("arquivo %q não existe no pacote", name)
	}
	return s.pkg.ReadFile(s.pkg.Files[idx])
}

func hashFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("calculando SHA-256 de %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func latestEntryModTime(files []ffxpackage.FileEntryV2) time.Time {
	var latest time.Time
	for i := range files {
		t := time.Unix(files[i].ModTime, 0)
		if t.After(latest) {
			latest = t
		}
	}
	return latest
}
