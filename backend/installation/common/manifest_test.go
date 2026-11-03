package common

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifestRoundTrip(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, ManifestFileName)

	modifiedAt := time.Date(2025, 3, 4, 5, 6, 7, 0, time.UTC)
	installedAt := time.Date(2025, 3, 5, 8, 9, 10, 0, time.UTC)
	want := &Manifest{
		SchemaVersion:  SchemaVersion,
		Product:        ProductID,
		DisplayName:    DisplayName,
		DisplayVersion: "2025.03.04",
		Status:         StatusInstalled,
		GameDir:        `C:\Games\FFX`,
		Package: PackageInfo{
			Path:          `resources\package.bin`,
			Name:          "package.bin",
			Size:          4096,
			SHA256:        "deadbeef",
			Magic:         "FFXX2HDR",
			FormatVersion: 2,
			FileCount:     2,
			ModifiedAt:    modifiedAt,
		},
		InstalledAt: installedAt,
		UpdatedAt:   installedAt,
		Uninstaller: UninstallerFileName,
		Files: []FileRecord{
			{Path: "data/mods/a.txt", Size: 5, SHA256: "aa", ModTime: modifiedAt, Scope: ScopeMods, Overwrote: true},
			{Path: "data/b.bin", Size: 7, SHA256: "bb", ModTime: modifiedAt, Scope: ScopeGame},
		},
		CreatedDirs: []string{"data/mods"},
	}

	if err := SaveManifest(path, want); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}

	got, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	if got.SchemaVersion != want.SchemaVersion ||
		got.Product != want.Product ||
		got.DisplayName != want.DisplayName ||
		got.DisplayVersion != want.DisplayVersion ||
		got.Status != want.Status ||
		got.GameDir != want.GameDir ||
		got.Uninstaller != want.Uninstaller {
		t.Fatalf("campos divergentes:\n got=%+v\nwant=%+v", got, want)
	}
	if !got.InstalledAt.Equal(want.InstalledAt) || !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Fatalf("datas divergentes: %v / %v", got.InstalledAt, got.UpdatedAt)
	}
	if got.Package != want.Package {
		t.Fatalf("pacote divergente: %+v", got.Package)
	}
	if len(got.Files) != len(want.Files) {
		t.Fatalf("arquivos divergentes: %+v", got.Files)
	}
	for i := range want.Files {
		if got.Files[i] != want.Files[i] {
			t.Fatalf("arquivo %d divergente: %+v", i, got.Files[i])
		}
	}
	if len(got.CreatedDirs) != 1 || got.CreatedDirs[0] != "data/mods" {
		t.Fatalf("diretórios divergentes: %+v", got.CreatedDirs)
	}
}

func TestManifestOverwritesAtomically(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, ManifestFileName)

	if err := SaveManifest(path, &Manifest{Status: StatusInstalling}); err != nil {
		t.Fatalf("primeiro save: %v", err)
	}
	if err := SaveManifest(path, &Manifest{Status: StatusInstalled}); err != nil {
		t.Fatalf("segundo save: %v", err)
	}

	got, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if got.Status != StatusInstalled {
		t.Fatalf("status = %q", got.Status)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != ManifestFileName {
			t.Fatalf("arquivo temporário deixado para trás: %s", entry.Name())
		}
	}
}

func TestManifestNotFound(t *testing.T) {
	_, err := LoadManifest(filepath.Join(tempDir(t), ManifestFileName))
	if !errors.Is(err, ErrManifestNotFound) {
		t.Fatalf("erro = %v, esperado ErrManifestNotFound", err)
	}
}

func TestManifestRejectsTamperedPayload(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, ManifestFileName)

	if err := SaveManifest(path, &Manifest{Status: StatusInstalled, GameDir: "x"}); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xFF
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = LoadManifest(path)
	if !errors.Is(err, ErrManifestCorrupt) {
		t.Fatalf("erro = %v, esperado ErrManifestCorrupt", err)
	}
}

func TestManifestRejectsBadMagic(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, ManifestFileName)

	if err := SaveManifest(path, &Manifest{Status: StatusInstalled}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[0] = 'X'
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = LoadManifest(path)
	if !errors.Is(err, ErrManifestCorrupt) {
		t.Fatalf("erro = %v, esperado ErrManifestCorrupt", err)
	}
}
