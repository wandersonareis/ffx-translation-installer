package common

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func installForUninstallTest(t *testing.T, gameDir string) *fakeRegistry {
	t.Helper()
	registry := &fakeRegistry{}
	installer := &Installer{
		Registry: registry,
		ExePath:  writeFakeExe(t, tempDir(t)),
	}
	source := newFakeSource(map[string]string{
		"data/mods/a.txt":     "hello",
		"data/mods/sub/b.bin": "world",
	}, time.Now())
	if _, err := installer.Install(source, gameDir); err != nil {
		t.Fatalf("Install: %v", err)
	}
	return registry
}

func TestUninstallRemovesOnlyRegisteredFiles(t *testing.T) {
	gameDir := tempDir(t)
	registry := installForUninstallTest(t, gameDir)

	userFile := filepath.Join(gameDir, "data", "mods", "user.txt")
	if err := os.WriteFile(userFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	uninstaller := &Uninstaller{Registry: registry, RunningExe: "/outro/exe"}
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods", "a.txt")); !os.IsNotExist(err) {
		t.Fatalf("a.txt deveria ter sido removido: %v", err)
	}
	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods", "sub", "b.bin")); !os.IsNotExist(err) {
		t.Fatalf("sub/b.bin deveria ter sido removido: %v", err)
	}
	if got, err := os.ReadFile(userFile); err != nil || string(got) != "keep" {
		t.Fatalf("arquivo do usuário não preservado: %q, %v", got, err)
	}
	if _, err := os.Stat(ManifestPath(gameDir)); !os.IsNotExist(err) {
		t.Fatalf("registro deveria ter sido removido: %v", err)
	}
	if _, err := os.Stat(UninstallerPath(gameDir)); !os.IsNotExist(err) {
		t.Fatalf("desinstalador deveria ter sido removido: %v", err)
	}
	if !result.UninstallerRemoved {
		t.Fatal("result.UninstallerRemoved = false")
	}
	if !result.RegistryRemoved {
		t.Fatal("result.RegistryRemoved = false")
	}
	if len(result.RemovedFiles) != 2 {
		t.Fatalf("RemovedFiles = %+v", result.RemovedFiles)
	}
	if len(result.RemovedDirs) != 1 || result.RemovedDirs[0] != "data/mods/sub" {
		t.Fatalf("RemovedDirs = %+v", result.RemovedDirs)
	}
	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods")); err != nil {
		t.Fatalf("data/mods deveria continuar existindo por causa do arquivo do usuário: %v", err)
	}
}

func TestUninstallRemovesCreatedDirsWhenEmpty(t *testing.T) {
	gameDir := tempDir(t)
	registry := installForUninstallTest(t, gameDir)

	uninstaller := &Uninstaller{Registry: registry, RunningExe: "/outro/exe"}
	if _, err := uninstaller.Uninstall(gameDir); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	if _, err := os.Stat(filepath.Join(gameDir, "data", "mods")); !os.IsNotExist(err) {
		t.Fatalf("data/mods deveria ter sido removido: %v", err)
	}
}

func TestUninstallSetsSelfDeletePathForRunningExe(t *testing.T) {
	gameDir := tempDir(t)
	registry := installForUninstallTest(t, gameDir)

	running := UninstallerPath(gameDir)
	uninstaller := &Uninstaller{Registry: registry, RunningExe: running}
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	if result.SelfDeletePath != running {
		t.Fatalf("SelfDeletePath = %q, esperado %q", result.SelfDeletePath, running)
	}
	if result.UninstallerRemoved {
		t.Fatal("UninstallerRemoved deveria ser false para o próprio exe")
	}
	if _, err := os.Stat(running); err != nil {
		t.Fatalf("desinstalador não deveria sumir aqui: %v", err)
	}
}

func TestUninstallMissingManifestIsNotAnError(t *testing.T) {
	gameDir := tempDir(t)
	registry := &fakeRegistry{}

	// Arquivos da tradução existem, mas sem o .dat não há como saber quais
	// são; eles precisam ser preservados.
	modsDir := filepath.Join(gameDir, "data", "mods")
	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	orphan := filepath.Join(modsDir, "traducao.txt")
	if err := os.WriteFile(orphan, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Um desinstalador copiado ainda existe; ele deve ser limpo também.
	if err := os.WriteFile(UninstallerPath(gameDir), []byte("exe"), 0o755); err != nil {
		t.Fatal(err)
	}

	uninstaller := &Uninstaller{Registry: registry, RunningExe: "/outro/exe"}
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.ManifestFound {
		t.Fatal("ManifestFound deveria ser false")
	}
	if len(result.Warnings) == 0 {
		t.Fatal("esperava aviso de registro não encontrado")
	}
	if !result.RegistryRemoved {
		t.Fatal("RegistryRemoved deveria ser true")
	}
	if !result.UninstallerRemoved {
		t.Fatal("UninstallerRemoved deveria ser true")
	}
	if _, err := os.Stat(UninstallerPath(gameDir)); !os.IsNotExist(err) {
		t.Fatalf("desinstalador deveria ter sido removido: %v", err)
	}
	if _, err := os.Stat(orphan); err != nil {
		t.Fatalf("arquivos em data/mods deveriam ser preservados: %v", err)
	}
	if _, err := os.Stat(ManifestPath(gameDir)); !os.IsNotExist(err) {
		t.Fatalf("não deveria existir .dat ao final: %v", err)
	}
}

func TestUninstallCorruptManifestIsNotAnError(t *testing.T) {
	gameDir := tempDir(t)
	path := ManifestPath(gameDir)
	if err := os.WriteFile(path, []byte("isto não é um registro"), 0o644); err != nil {
		t.Fatal(err)
	}

	modsDir := filepath.Join(gameDir, "data", "mods")
	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	orphan := filepath.Join(modsDir, "traducao.txt")
	if err := os.WriteFile(orphan, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	registry := &fakeRegistry{}
	uninstaller := &Uninstaller{Registry: registry, RunningExe: "/outro/exe"}
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.ManifestFound {
		t.Fatal("ManifestFound deveria ser false")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("registro corrompido deveria ter sido removido: %v", err)
	}
	if !result.RegistryRemoved {
		t.Fatal("RegistryRemoved deveria ser true")
	}
	if _, err := os.Stat(orphan); err != nil {
		t.Fatalf("arquivos em data/mods deveriam ser preservados: %v", err)
	}
}

func TestUninstallSkipsUnsafePathFromManifest(t *testing.T) {
	gameDir := tempDir(t)
	outside := filepath.Join(filepath.Dir(gameDir), "evil.txt")
	if err := os.WriteFile(outside, []byte("do not delete"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(outside) })

	manifest := &Manifest{
		SchemaVersion: SchemaVersion,
		Status:        StatusInstalled,
		GameDir:       gameDir,
		Files: []FileRecord{
			{Path: "../evil.txt", Scope: ScopeGame},
		},
	}
	if err := SaveManifest(ManifestPath(gameDir), manifest); err != nil {
		t.Fatal(err)
	}

	uninstaller := &Uninstaller{Registry: &fakeRegistry{}}
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("esperava aviso para caminho inválido")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("arquivo fora da pasta do jogo foi removido: %v", err)
	}
}

func TestUninstallIgnoresMissingRecordedFile(t *testing.T) {
	gameDir := tempDir(t)
	registry := installForUninstallTest(t, gameDir)

	if err := os.Remove(filepath.Join(gameDir, "data", "mods", "a.txt")); err != nil {
		t.Fatal(err)
	}

	uninstaller := &Uninstaller{Registry: registry, RunningExe: "/outro/exe"}
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if len(result.RemovedFiles) != 1 {
		t.Fatalf("RemovedFiles = %+v", result.RemovedFiles)
	}
}
