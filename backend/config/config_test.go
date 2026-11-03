package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ffxinstaller/backend/common"
)

func newTestConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()
	return &Config{
		filePath: filepath.Join(dir, "config", "config.json"),
	}
}

func TestSetGameExeDirPersists(t *testing.T) {
	c := newTestConfig(t)
	gameDir := t.TempDir()

	if err := c.SetGameExeDir("  " + gameDir + "  "); err != nil {
		t.Fatalf("set válido: %v", err)
	}
	if c.GameExeDir != gameDir {
		t.Fatalf("caminho não normalizado: %q", c.GameExeDir)
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		t.Fatalf("ler config: %v", err)
	}
	var persisted map[string]string
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if persisted["GameExeDir"] != gameDir {
		t.Fatalf("GameExeDir não persistido: %s", data)
	}
	if _, ok := persisted["GameExeFile"]; ok {
		t.Fatalf("GameExeFile não deveria mais existir no config: %s", data)
	}
	if _, ok := persisted["ResourcesDir"]; ok {
		t.Fatalf("ResourcesDir não deveria ser persistido (é fixo): %s", data)
	}

	if err := c.SetGameExeDir(filepath.Join(gameDir, "nao-existe")); err == nil {
		t.Fatal("diretório inexistente deveria ser recusado")
	}
	if c.GameExeDir != gameDir {
		t.Fatalf("recusa alterou o valor persistido: %q", c.GameExeDir)
	}

	if err := c.SetGameExeDir(""); err != nil {
		t.Fatalf("limpar deveria ser válido: %v", err)
	}
	if c.GameExeDir != "" {
		t.Fatalf("limpar não zerou: %q", c.GameExeDir)
	}
}

func TestSnapshotIncludesDerivedPackageBin(t *testing.T) {
	c := newTestConfig(t)
	snap := c.Snapshot()
	want := filepath.Join(common.GetExecDir(), "resources", "package.bin")
	if snap.PackageBinPath != want {
		t.Fatalf("PackageBinPath = %q, esperado %q", snap.PackageBinPath, want)
	}
	if snap.ResourcesDir != filepath.Dir(want) {
		t.Fatalf("ResourcesDir = %q, esperado %q", snap.ResourcesDir, filepath.Dir(want))
	}
}

// O diretório de recursos é fixo: <execDir>/resources, sempre derivado.
func TestResourcesDirIsDerived(t *testing.T) {
	c := newTestConfig(t)
	want := filepath.Join(common.GetExecDir(), "resources")
	if got := c.GetResourcesDir(); got != want {
		t.Fatalf("GetResourcesDir = %q, esperado %q", got, want)
	}
	if err := c.EnsureResourcesDir(); err != nil {
		t.Fatalf("EnsureResourcesDir: %v", err)
	}
	if info, err := os.Stat(want); err != nil || !info.IsDir() {
		t.Fatalf("diretório de recursos não foi criado: %v", err)
	}
}

// Sem conseguir preparar o diretório de recursos, o erro precisa subir para a
// UI (o app não tem onde gravar o package.bin).
func TestEnsureResourcesDirFailsWhenPathIsFile(t *testing.T) {
	c := newTestConfig(t)
	resources := c.GetResourcesDir()

	// Preserva um diretório já criado por outra execução.
	backup := resources + ".test-bak"
	if info, err := os.Stat(resources); err == nil && info.IsDir() {
		if err := os.Rename(resources, backup); err != nil {
			t.Fatalf("preparar: %v", err)
		}
		defer func() { _ = os.Rename(backup, resources) }()
	}

	if err := os.WriteFile(resources, []byte("arquivo, não pasta"), 0o644); err != nil {
		t.Fatalf("criar arquivo: %v", err)
	}
	defer func() { _ = os.Remove(resources) }()

	if err := c.EnsureResourcesDir(); err == nil {
		t.Fatal("esperava erro ao preparar resources como arquivo")
	}
}

func TestLoadKeepsMissingGameExeDir(t *testing.T) {
	c := newTestConfig(t)
	if err := os.MkdirAll(filepath.Dir(c.filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(t.TempDir(), "drive-offline", "FFX")
	content, err := json.Marshal(map[string]string{
		"GameExeDir": missing,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.filePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := c.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.GameExeDir != missing {
		t.Fatalf("GameExeDir foi apagado no boot: %q", c.GameExeDir)
	}
}

// Configs antigas guardavam o caminho completo do .exe; o valor precisa ser
// migrado para o diretório e o campo legado removido do arquivo.
func TestLoadMigratesLegacyGameExeFile(t *testing.T) {
	c := newTestConfig(t)
	if err := os.MkdirAll(filepath.Dir(c.filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	wantDir := t.TempDir()
	content, err := json.Marshal(map[string]string{
		"GameExeFile": filepath.Join(wantDir, "FFX-2.exe"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.filePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := c.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.GameExeDir != wantDir {
		t.Fatalf("GameExeDir = %q, esperado %q (migrado do GameExeFile)", c.GameExeDir, wantDir)
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		t.Fatalf("ler config: %v", err)
	}
	var persisted map[string]string
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if _, ok := persisted["GameExeFile"]; ok {
		t.Fatalf("campo legado GameExeFile deveria sair do JSON: %s", data)
	}
}

// Se o config já tem o diretório, o campo legado não o sobrescreve.
func TestLoadKeepsGameExeDirOverLegacyFile(t *testing.T) {
	c := newTestConfig(t)
	if err := os.MkdirAll(filepath.Dir(c.filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	currentDir := t.TempDir()
	content, err := json.Marshal(map[string]string{
		"GameExeDir":  currentDir,
		"GameExeFile": filepath.Join(t.TempDir(), "FFX.exe"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.filePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := c.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.GameExeDir != currentDir {
		t.Fatalf("GameExeDir = %q, esperado %q", c.GameExeDir, currentDir)
	}
}

func TestThemeDefaultsToSystem(t *testing.T) {
	c := newTestConfig(t)
	if got := c.GetTheme(); got != "system" {
		t.Fatalf("GetTheme = %q, esperado system", got)
	}
	if got := c.Snapshot().Theme; got != "system" {
		t.Fatalf("Snapshot.Theme = %q, esperado system", got)
	}
}

func TestSetThemePersists(t *testing.T) {
	c := newTestConfig(t)

	if err := c.SetTheme(" LIGHT "); err != nil {
		t.Fatalf("set theme: %v", err)
	}
	if c.Theme != "light" {
		t.Fatalf("tema não normalizado: %q", c.Theme)
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		t.Fatalf("ler config: %v", err)
	}
	if !strings.Contains(string(data), `"Theme": "light"`) {
		t.Fatalf("Theme não persistido: %s", data)
	}

	if err := c.SetTheme("blue"); err == nil {
		t.Fatal("tema inválido deveria ser recusado")
	}
	if c.Theme != "light" {
		t.Fatalf("recusa alterou o tema persistido: %q", c.Theme)
	}

	if err := c.SetTheme("system"); err != nil {
		t.Fatalf("system deveria ser aceito: %v", err)
	}
	if got := c.Snapshot().Theme; got != "system" {
		t.Fatalf("Snapshot.Theme = %q, esperado system", got)
	}
}

func TestLoadToleratesUTF8BOM(t *testing.T) {
	c := newTestConfig(t)
	if err := os.MkdirAll(filepath.Dir(c.filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	content, err := json.Marshal(map[string]string{
		"Theme": "dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, content...)
	if err := os.WriteFile(c.filePath, withBOM, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Load(); err != nil {
		t.Fatalf("Load com BOM: %v", err)
	}
	if c.Theme != "dark" {
		t.Fatalf("config com BOM não foi lida: %+v", c)
	}
}

func TestLoadNormalizesTheme(t *testing.T) {
	c := newTestConfig(t)
	if err := os.MkdirAll(filepath.Dir(c.filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	content, err := json.Marshal(map[string]string{
		"Theme": " DARK ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.filePath, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Theme != "dark" || c.Snapshot().Theme != "dark" {
		t.Fatalf("tema não normalizado no Load: %q", c.Theme)
	}

	invalid, err := json.Marshal(map[string]string{
		"Theme": "blue",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.filePath, invalid, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Theme != "" || c.Snapshot().Theme != "system" {
		t.Fatalf("tema inválido deveria voltar para system: %q", c.Theme)
	}
}
