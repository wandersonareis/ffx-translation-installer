package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInstallWritesPackageAndManifest(t *testing.T) {
	gameDir := tempDir(t)
	exe := writeFakeExe(t, tempDir(t))
	registry := &fakeRegistry{}
	installer := &Installer{Registry: registry, ExePath: exe}

	modTime := time.Date(2025, 3, 4, 5, 6, 7, 0, time.UTC)
	source := newFakeSource(map[string]string{
		"data/mods/a.txt":     "hello",
		"data/mods/sub/b.bin": "world",
	}, modTime)

	result, err := installer.Install(source, gameDir)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	manifest := result.Manifest
	if manifest == nil {
		t.Fatal("manifest nil")
	}
	if manifest.Status != StatusInstalled {
		t.Fatalf("status = %q", manifest.Status)
	}
	if manifest.GameDir != gameDir {
		t.Fatalf("GameDir = %q, esperado %q", manifest.GameDir, gameDir)
	}
	if manifest.DisplayVersion != "2025.03.04" {
		t.Fatalf("DisplayVersion = %q", manifest.DisplayVersion)
	}
	if manifest.Package.SHA256 != "abc123" || manifest.Package.Magic != "FFXX2HDR" {
		t.Fatalf("pacote = %+v", manifest.Package)
	}

	checks := map[string]string{
		filepath.Join("data", "mods", "a.txt"):        "hello",
		filepath.Join("data", "mods", "sub", "b.bin"): "world",
	}
	for rel, want := range checks {
		got, err := os.ReadFile(filepath.Join(gameDir, rel))
		if err != nil {
			t.Fatalf("lendo %s: %v", rel, err)
		}
		if string(got) != want {
			t.Fatalf("%s = %q, esperado %q", rel, got, want)
		}
	}

	info, err := os.Stat(filepath.Join(gameDir, "data", "mods", "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.ModTime().Unix() != modTime.Unix() {
		t.Fatalf("modtime = %v, esperado %v", info.ModTime(), modTime)
	}

	loaded, err := LoadManifest(ManifestPath(gameDir))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(loaded.Files) != 2 {
		t.Fatalf("Files = %+v", loaded.Files)
	}
	for _, file := range loaded.Files {
		if file.Scope != ScopeMods {
			t.Fatalf("Scope de %s = %q", file.Path, file.Scope)
		}
		if file.Size == 0 || file.SHA256 == "" {
			t.Fatalf("arquivo sem tamanho/hash: %+v", file)
		}
	}

	if _, err := os.Stat(UninstallerPath(gameDir)); err != nil {
		t.Fatalf("desinstalador não copiado: %v", err)
	}
	if len(registry.entries) != 1 {
		t.Fatalf("entradas no registro = %d", len(registry.entries))
	}
	entry := registry.entries[0]
	if entry.DisplayName != DisplayName {
		t.Fatalf("DisplayName = %q", entry.DisplayName)
	}
	if entry.InstallLocation != manifest.GameDir {
		t.Fatalf("InstallLocation = %q", entry.InstallLocation)
	}
	if entry.UninstallString == "" || entry.QuietUninstallString == "" {
		t.Fatalf("strings de desinstalação vazias: %+v", entry)
	}
}

func TestInstallOverwritesPackageFileAndPreservesUserFile(t *testing.T) {
	gameDir := tempDir(t)
	exe := writeFakeExe(t, tempDir(t))
	installer := &Installer{Registry: &fakeRegistry{}, ExePath: exe}

	modsDir := filepath.Join(gameDir, "data", "mods")
	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modsDir, "a.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modsDir, "user.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	source := newFakeSource(map[string]string{"data/mods/a.txt": "new"}, time.Now())
	result, err := installer.Install(source, gameDir)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !result.Manifest.Files[0].Overwrote {
		t.Fatal("Overwrote deveria ser true")
	}

	got, err := os.ReadFile(filepath.Join(modsDir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("a.txt = %q", got)
	}
	if got, err := os.ReadFile(filepath.Join(modsDir, "user.txt")); err != nil || string(got) != "keep" {
		t.Fatalf("arquivo do usuário alterado: %q, %v", got, err)
	}
}

func TestInstallReplacesPreviousInstallation(t *testing.T) {
	gameDir := tempDir(t)
	exe := writeFakeExe(t, tempDir(t))
	registry := &fakeRegistry{}
	installer := &Installer{Registry: registry, ExePath: exe}

	first := newFakeSource(map[string]string{
		"data/mods/a.txt":   "v1",
		"data/mods/old.txt": "old",
	}, time.Now())
	if _, err := installer.Install(first, gameDir); err != nil {
		t.Fatalf("primeira instalação: %v", err)
	}

	second := newFakeSource(map[string]string{
		"data/mods/a.txt":   "v2",
		"data/mods/new.txt": "new",
	}, time.Now())
	result, err := installer.Install(second, gameDir)
	if err != nil {
		t.Fatalf("segunda instalação: %v", err)
	}

	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods", "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("old.txt deveria ter sido removido: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(gameDir, "data", "mods", "a.txt")); err != nil || string(got) != "v2" {
		t.Fatalf("a.txt = %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods", "new.txt")); err != nil {
		t.Fatalf("new.txt não foi criado: %v", err)
	}
	if len(result.Manifest.Files) != 2 {
		t.Fatalf("Files do novo registro = %+v", result.Manifest.Files)
	}
	if registry.removes != 1 {
		t.Fatalf("Remove do registro chamado %d vez(es)", registry.removes)
	}
}

func TestInstallRollsBackOnFailure(t *testing.T) {
	gameDir := tempDir(t)
	exe := writeFakeExe(t, tempDir(t))
	registry := &fakeRegistry{}
	installer := &Installer{Registry: registry, ExePath: exe}

	// z.txt é um diretório: a escrita falha depois que a.txt já foi gravado.
	if err := os.MkdirAll(filepath.Join(gameDir, "data", "mods", "z.txt"), 0o755); err != nil {
		t.Fatal(err)
	}

	source := newFakeSource(map[string]string{
		"data/mods/a.txt": "written",
		"data/mods/z.txt": "fails",
	}, time.Now())

	if _, err := installer.Install(source, gameDir); err == nil {
		t.Fatal("Install deveria falhar")
	}

	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods", "a.txt")); !os.IsNotExist(err) {
		t.Fatalf("a.txt deveria ter sido removido no rollback: %v", err)
	}
	if _, err := os.Stat(ManifestPath(gameDir)); !os.IsNotExist(err) {
		t.Fatalf("registro deveria ter sido removido: %v", err)
	}
	if _, err := os.Stat(UninstallerPath(gameDir)); !os.IsNotExist(err) {
		t.Fatalf("desinstalador deveria ter sido removido: %v", err)
	}
	if registry.removes == 0 {
		t.Fatal("Remove do registro não foi chamado no rollback")
	}
}

func TestInstallRollsBackWhenRegistryFails(t *testing.T) {
	gameDir := tempDir(t)
	exe := writeFakeExe(t, tempDir(t))
	registry := &fakeRegistry{failNext: true}
	installer := &Installer{Registry: registry, ExePath: exe}

	source := newFakeSource(map[string]string{"data/mods/a.txt": "x"}, time.Now())
	if _, err := installer.Install(source, gameDir); err == nil {
		t.Fatal("Install deveria falhar quando o registro falha")
	}

	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods", "a.txt")); !os.IsNotExist(err) {
		t.Fatalf("a.txt deveria ter sido removido: %v", err)
	}
	if _, err := os.Stat(ManifestPath(gameDir)); !os.IsNotExist(err) {
		t.Fatalf("registro deveria ter sido removido: %v", err)
	}
}

func TestInstallRejectsDuplicatePaths(t *testing.T) {
	gameDir := tempDir(t)
	installer := &Installer{Registry: &fakeRegistry{}, ExePath: writeFakeExe(t, tempDir(t))}

	source := &fakeSource{
		info: PackageInfo{Magic: "FFXX2HDR", FormatVersion: 2},
		entries: []SourceEntry{
			{Name: "data/mods/a.txt"},
			{Name: `data\mods\a.txt`},
		},
		data: map[string][]byte{
			"data/mods/a.txt": []byte("x"),
			`data\mods\a.txt`: []byte("x"),
		},
	}

	_, err := installer.Install(source, gameDir)
	if err == nil || !strings.Contains(err.Error(), "duplicados") {
		t.Fatalf("erro = %v, esperado caminhos duplicados", err)
	}
}

func TestInstallRejectsUnsafePackagePath(t *testing.T) {
	gameDir := tempDir(t)
	installer := &Installer{Registry: &fakeRegistry{}, ExePath: writeFakeExe(t, tempDir(t))}

	source := newFakeSource(map[string]string{"../evil.txt": "x"}, time.Now())
	if _, err := installer.Install(source, gameDir); err == nil {
		t.Fatal("Install deveria recusar caminho fora da pasta do jogo")
	}
}

func TestInstallRejectsEmptyPackage(t *testing.T) {
	gameDir := tempDir(t)
	installer := &Installer{Registry: &fakeRegistry{}, ExePath: writeFakeExe(t, tempDir(t))}

	source := newFakeSource(map[string]string{}, time.Now())
	if _, err := installer.Install(source, gameDir); err == nil {
		t.Fatal("Install deveria recusar pacote vazio")
	}
}

func TestInstallRejectsMissingGameDir(t *testing.T) {
	installer := &Installer{Registry: &fakeRegistry{}, ExePath: writeFakeExe(t, tempDir(t))}
	source := newFakeSource(map[string]string{"data/mods/a.txt": "x"}, time.Now())

	if _, err := installer.Install(source, filepath.Join(tempDir(t), "nao-existe")); err == nil {
		t.Fatal("Install deveria recusar pasta inexistente")
	}
}
