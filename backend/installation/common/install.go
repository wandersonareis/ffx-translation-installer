package common

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ffxinstaller/backend/installation/system"
)

// Installer executa a instalação: copia o pacote, grava o registro binário,
// publica o desinstalador e registra a entrada de desinstalação do sistema.
type Installer struct {
	// Registry é o registro do sistema (injetável nos testes).
	Registry system.UninstallRegistry
	// ExePath é o executável copiado para a pasta do jogo como desinstalador.
	ExePath string
}

// InstallResult devolve o registro final e avisos não fatais.
type InstallResult struct {
	Manifest *Manifest `json:"Manifest"`
	Warnings []string  `json:"Warnings,omitempty"`
}

// NewInstaller cria um instalador usando o registro do sistema e o executável
// atual como desinstalador.
func NewInstaller() *Installer {
	exe, _ := os.Executable()
	return &Installer{
		Registry: system.NewUninstallRegistry(ProductID),
		ExePath:  exe,
	}
}

type plannedFile struct {
	rel   string
	entry SourceEntry
}

// Install copia todas as entradas do pacote para gameDir e grava o registro.
//
// Uma instalação anterior na mesma pasta é removida antes (somente os arquivos
// registrados por ela). Em qualquer erro, o que esta instalação criou é
// desfeito.
func (i *Installer) Install(src Source, gameDir string) (*InstallResult, error) {
	gameDir = strings.TrimSpace(gameDir)
	if gameDir == "" {
		return nil, fmt.Errorf("diretório do jogo não informado")
	}
	absGameDir, err := filepath.Abs(gameDir)
	if err != nil {
		return nil, fmt.Errorf("resolvendo diretório do jogo: %w", err)
	}
	if info, err := os.Stat(absGameDir); err != nil {
		return nil, fmt.Errorf("pasta do jogo não encontrada: %s: %w", absGameDir, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("o caminho do jogo não é uma pasta: %s", absGameDir)
	}

	if src == nil {
		return nil, fmt.Errorf("pacote não informado")
	}
	pkg := src.Info()
	entries := src.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("pacote vazio: nenhum arquivo para instalar")
	}

	// Valida todos os caminhos ANTES de tocar em qualquer arquivo.
	planned := make([]plannedFile, 0, len(entries))
	seen := make(map[string]string, len(entries))
	for _, entry := range entries {
		rel, err := NormalizeRelPath(entry.Name)
		if err != nil {
			return nil, err
		}
		key := strings.ToLower(rel)
		if other, ok := seen[key]; ok {
			return nil, fmt.Errorf("caminhos duplicados no pacote: %q e %q", other, entry.Name)
		}
		seen[key] = entry.Name
		planned = append(planned, plannedFile{rel: rel, entry: entry})
	}

	// Pré-valida a integridade do pacote (ReadFile confere CRC32 e xxHash64)
	// antes de remover a instalação anterior.
	records := make([]FileRecord, 0, len(planned))
	for _, plan := range planned {
		data, err := src.ReadFile(plan.entry.Name)
		if err != nil {
			return nil, fmt.Errorf("verificando %s: %w", plan.entry.Name, err)
		}
		records = append(records, FileRecord{
			Path:    plan.rel,
			Size:    int64(len(data)),
			SHA256:  sha256Hex(data),
			ModTime: plan.entry.ModTime,
			Scope:   scopeOf(plan.rel),
		})
	}

	result := &InstallResult{}

	// Instalar substitui a instalação anterior da mesma pasta do jogo.
	previous, loadErr := LoadManifest(ManifestPath(absGameDir))
	switch {
	case loadErr == nil && previous != nil:
		uninstaller := &Uninstaller{Registry: i.Registry, RunningExe: i.ExePath}
		removed, err := uninstaller.Uninstall(absGameDir)
		if err != nil {
			return nil, fmt.Errorf("removendo instalação anterior: %w", err)
		}
		if removed != nil && removed.FailedFiles > 0 {
			return nil, fmt.Errorf(
				"não foi possível remover a instalação anterior: %d arquivo(s) em uso; feche o jogo e tente novamente",
				removed.FailedFiles,
			)
		}
	case errors.Is(loadErr, ErrManifestNotFound):
		// Primeira instalação.
	default:
		result.Warnings = append(result.Warnings,
			"registro anterior inválido; os arquivos da instalação anterior podem ter permanecido")
	}

	now := time.Now().UTC()
	manifest := &Manifest{
		SchemaVersion: SchemaVersion,
		Product:       ProductID,
		DisplayName:   DisplayName,
		Status:        StatusInstalling,
		GameDir:       absGameDir,
		Package:       pkg,
		InstalledAt:   now,
		UpdatedAt:     now,
		Uninstaller:   UninstallerFileName,
		Files:         records,
	}
	manifest.DisplayVersion = translationVersion(manifest)

	if err := SaveManifest(ManifestPath(absGameDir), manifest); err != nil {
		return nil, fmt.Errorf("gravando registro de instalação: %w", err)
	}

	written := make([]string, 0, len(records))
	createdDirs := make([]string, 0)

	rollback := func() {
		for idx := len(written) - 1; idx >= 0; idx-- {
			_ = os.Remove(written[idx])
		}
		removeEmptyDirs(absGameDir, createdDirs)
		_ = os.Remove(UninstallerPath(absGameDir))
		_ = i.Registry.Remove()
		_ = os.Remove(ManifestPath(absGameDir))
	}

	for idx := range records {
		record := &records[idx]
		target := filepath.Join(absGameDir, filepath.FromSlash(record.Path))

		dirs, err := ensureParentDirs(absGameDir, target)
		createdDirs = append(createdDirs, dirs...)
		if err != nil {
			rollback()
			return nil, err
		}

		if info, statErr := os.Stat(target); statErr == nil && !info.IsDir() {
			record.Overwrote = true
		}

		data, err := src.ReadFile(planned[idx].entry.Name)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("lendo %s: %w", record.Path, err)
		}
		if sha256Hex(data) != record.SHA256 {
			rollback()
			return nil, fmt.Errorf("o pacote mudou durante a instalação: %s", record.Path)
		}

		if err := writeFileAtomic(target, data, 0o644, record.ModTime); err != nil {
			rollback()
			return nil, fmt.Errorf("gravando %s: %w", record.Path, err)
		}
		written = append(written, target)
	}

	if err := i.deployUninstaller(absGameDir); err != nil {
		rollback()
		return nil, err
	}

	manifest.CreatedDirs = createdDirs
	manifest.Files = records
	manifest.Status = StatusInstalled
	manifest.UpdatedAt = time.Now().UTC()

	if err := i.Registry.Register(uninstallEntry(manifest)); err != nil {
		rollback()
		return nil, err
	}

	if err := SaveManifest(ManifestPath(absGameDir), manifest); err != nil {
		rollback()
		return nil, fmt.Errorf("finalizando registro de instalação: %w", err)
	}

	result.Manifest = manifest
	return result, nil
}

func (i *Installer) deployUninstaller(gameDir string) error {
	if strings.TrimSpace(i.ExePath) == "" {
		return fmt.Errorf("executável do instalador não informado")
	}
	destination := UninstallerPath(gameDir)
	if samePath(i.ExePath, destination) {
		return nil
	}
	if err := copyFileAtomic(i.ExePath, destination, 0o755); err != nil {
		return fmt.Errorf("copiando desinstalador: %w", err)
	}
	return nil
}

func scopeOf(rel string) string {
	lower := strings.ToLower(rel)
	if lower == "data/mods" || strings.HasPrefix(lower, "data/mods/") {
		return ScopeMods
	}
	return ScopeGame
}

func translationVersion(m *Manifest) string {
	t := m.Package.ModifiedAt
	if t.IsZero() {
		t = m.InstalledAt
	}
	return t.Local().Format("2006.01.02")
}

func uninstallEntry(m *Manifest) system.UninstallEntry {
	exe := UninstallerPath(m.GameDir)

	var total int64
	for _, file := range m.Files {
		total += file.Size
	}
	var sizeKB uint32
	if total > 0 {
		kb := (total + 1023) / 1024
		if kb > math.MaxUint32 {
			kb = math.MaxUint32
		}
		sizeKB = uint32(kb)
	}

	installedAt := m.InstalledAt
	if installedAt.IsZero() {
		installedAt = time.Now()
	}

	return system.UninstallEntry{
		DisplayName:          m.DisplayName,
		DisplayVersion:       m.DisplayVersion,
		DisplayIcon:          quote(exe) + ",0",
		InstallLocation:      m.GameDir,
		UninstallString:      quote(exe),
		QuietUninstallString: quote(exe) + " " + quietFlag,
		InstallDate:          installedAt.Local().Format("20060102"),
		EstimatedSizeKB:      sizeKB,
	}
}

func quote(path string) string {
	return `"` + path + `"`
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
