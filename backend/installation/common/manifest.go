// Package common registra e remove instalações da tradução.
//
// O registro fica na pasta do jogo, ao lado do desinstalador, num arquivo
// binário (gob) com magic, versão e checksum SHA-256. Não é texto editável:
// se o conteúdo for alterado/corrompido, o arquivo é tratado como inexistente.
package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// ManifestFileName é o registro binário gravado na pasta do jogo.
	ManifestFileName = "ffx-translation-uninstall.dat"
	// UninstallerFileName é a cópia do instalador usada como desinstalador.
	UninstallerFileName = "ffx-translation-uninstall.exe"

	// ProductID identifica a tradução no registro do Windows.
	ProductID = "FFXTranslation"
	// DisplayName é o nome exibido em "Aplicativos e recursos".
	DisplayName = "Tradução FINAL FANTASY X/X-2 HD Remaster"

	// SchemaVersion é a versão do formato do registro binário.
	SchemaVersion = 1

	// StatusInstalling marca uma instalação em andamento (recuperação de crash).
	StatusInstalling = "installing"
	// StatusInstalled marca uma instalação concluída.
	StatusInstalled = "installed"

	// ScopeMods identifica arquivos dentro de data/mods.
	ScopeMods = "mods"
	// ScopeGame identifica arquivos fora de data/mods.
	ScopeGame = "game"

	// manifestHeaderSize = magic (8) + versão (1) + tamanho (4) + SHA-256 (32).
	manifestHeaderSize = 8 + 1 + 4 + sha256.Size
	// maxManifestSize evita alocar memória demais com um arquivo corrompido.
	maxManifestSize = 64 << 20
)

var manifestMagic = [8]byte{'F', 'F', 'X', 'I', 'M', 'A', 'N', 0}

var (
	// ErrManifestNotFound indica que o registro não existe.
	ErrManifestNotFound = errors.New("registro de instalação não encontrado")
	// ErrManifestCorrupt indica que o registro existe mas não pôde ser validado.
	ErrManifestCorrupt = errors.New("registro de instalação inválido")
)

// Manifest é o registro completo de uma instalação.
type Manifest struct {
	SchemaVersion  int          `json:"SchemaVersion"`
	Product        string       `json:"Product"`
	DisplayName    string       `json:"DisplayName"`
	DisplayVersion string       `json:"DisplayVersion"`
	Status         string       `json:"Status"`
	GameDir        string       `json:"GameDir"`
	Package        PackageInfo  `json:"Package"`
	InstalledAt    time.Time    `json:"InstalledAt"`
	UpdatedAt      time.Time    `json:"UpdatedAt"`
	Uninstaller    string       `json:"Uninstaller"`
	Files          []FileRecord `json:"Files"`
	CreatedDirs    []string     `json:"CreatedDirs"`
}

// PackageInfo guarda os dados úteis do pacote instalado.
type PackageInfo struct {
	Path          string `json:"Path"`
	Name          string `json:"Name"`
	Size          int64  `json:"Size"`
	SHA256        string `json:"SHA256"`
	Magic         string `json:"Magic"`
	FormatVersion uint32 `json:"FormatVersion"`
	FileCount     uint32 `json:"FileCount"`
	// ModifiedAt é a data da tradução: modtime do pacote (ou da entrada mais
	// recente quando o sistema de arquivos não informa um valor válido).
	ModifiedAt time.Time `json:"ModifiedAt"`
}

// FileRecord descreve um arquivo copiado do pacote.
type FileRecord struct {
	// Path é relativo à pasta do jogo e usa "/" como separador.
	Path    string    `json:"Path"`
	Size    int64     `json:"Size"`
	SHA256  string    `json:"SHA256"`
	ModTime time.Time `json:"ModTime"`
	Scope   string    `json:"Scope"`
	// Overwrote indica que já existia um arquivo no destino.
	Overwrote bool `json:"Overwrote"`
}

// ManifestPath devolve o caminho do registro na pasta do jogo.
func ManifestPath(gameDir string) string {
	return filepath.Join(gameDir, ManifestFileName)
}

// UninstallerPath devolve o caminho do desinstalador na pasta do jogo.
func UninstallerPath(gameDir string) string {
	return filepath.Join(gameDir, UninstallerFileName)
}

// SaveManifest grava o registro de forma atômica.
func SaveManifest(path string, m *Manifest) error {
	if m == nil {
		return fmt.Errorf("registro vazio")
	}
	if m.SchemaVersion == 0 {
		m.SchemaVersion = SchemaVersion
	}
	if m.SchemaVersion != SchemaVersion {
		return fmt.Errorf("versão de registro não suportada: %d", m.SchemaVersion)
	}

	var payload bytes.Buffer
	if err := gob.NewEncoder(&payload).Encode(m); err != nil {
		return fmt.Errorf("codificando registro: %w", err)
	}
	if payload.Len() > maxManifestSize {
		return fmt.Errorf("registro grande demais: %d bytes", payload.Len())
	}

	sum := sha256.Sum256(payload.Bytes())

	var out bytes.Buffer
	out.Write(manifestMagic[:])
	out.WriteByte(byte(SchemaVersion))
	var size [4]byte
	binary.LittleEndian.PutUint32(size[:], uint32(payload.Len()))
	out.Write(size[:])
	out.Write(sum[:])
	out.Write(payload.Bytes())

	return writeFileAtomic(path, out.Bytes(), 0o644, time.Time{})
}

// LoadManifest lê e valida o registro.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrManifestNotFound
		}
		return nil, err
	}

	if len(data) < manifestHeaderSize {
		return nil, fmt.Errorf("%w: arquivo truncado", ErrManifestCorrupt)
	}
	if !bytes.Equal(data[:8], manifestMagic[:]) {
		return nil, fmt.Errorf("%w: magic inválido", ErrManifestCorrupt)
	}
	if version := data[8]; version != SchemaVersion {
		return nil, fmt.Errorf("%w: versão %d não suportada", ErrManifestCorrupt, version)
	}

	size := binary.LittleEndian.Uint32(data[9:13])
	if size > maxManifestSize || len(data) != manifestHeaderSize+int(size) {
		return nil, fmt.Errorf("%w: tamanho inválido", ErrManifestCorrupt)
	}

	payload := data[manifestHeaderSize:]
	sum := sha256.Sum256(payload)
	if !bytes.Equal(data[13:manifestHeaderSize], sum[:]) {
		return nil, fmt.Errorf("%w: checksum não confere", ErrManifestCorrupt)
	}

	var m Manifest
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrManifestCorrupt, err)
	}
	if m.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("%w: versão interna %d", ErrManifestCorrupt, m.SchemaVersion)
	}
	return &m, nil
}

// writeFileAtomic grava data num arquivo temporário no mesmo diretório e
// renomeia por cima do destino. Em erro, o temporário é removido.
func writeFileAtomic(path string, data []byte, perm os.FileMode, modTime time.Time) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ffx-install-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if !modTime.IsZero() {
		if err := os.Chtimes(tmpName, modTime, modTime); err != nil {
			_ = os.Remove(tmpName)
			return err
		}
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}
