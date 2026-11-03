// Package config gerencia a configuração persistente do instalador.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ffxinstaller/backend/common"
)

// Config armazena a configuração do instalador.
type Config struct {
	filePath string

	// GameExeDir é o diretório do jogo (onde estão FFX.exe/FFX-2.exe e os
	// data/*.vbf). É o ÚNICO caminho do jogo: o seletor pede um dos .exe só
	// para garantir a pasta certa, mas persiste apenas o diretório — o mesmo
	// formato do caminho devolvido pelo serviço que procura o jogo no Steam.
	GameExeDir string `json:"GameExeDir"`

	// Theme é a preferência visual: "" (ou "system") segue o tema do
	// sistema operacional; "light"/"dark" são escolhas explícitas.
	Theme string `json:"Theme,omitempty"`
}

// ConfigSnapshot é o espelho enviado ao frontend: inclui os caminhos
// derivados (não persistidos, para não envelhecerem no JSON) e o tema efetivo
// ("system" quando o config.json não tem escolha salva).
type ConfigSnapshot struct {
	ResourcesDir   string `json:"ResourcesDir"`
	GameExeDir     string `json:"GameExeDir"`
	PackageBinPath string `json:"PackageBinPath"`
	Theme          string `json:"Theme"`
}

// NewConfig cria uma nova instância de configuração.
func NewConfig() *Config {
	filePath := filepath.Join(common.GetExecDir(), "config", "config.json")

	c := &Config{
		filePath: filePath,
	}

	if err := c.Load(); err != nil {
		common.LogVerbose("Using default config (%v)", err)
	}

	return c
}

// Load carrega a configuração do arquivo JSON.
func (c *Config) Load() error {
	file, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.saveDefaults()
		}
		return err
	}

	// Notepad/PowerShell podem gravar UTF-8 com BOM: o decoder JSON não
	// aceita o prefixo, então ele é removido antes do parse.
	file = bytes.TrimPrefix(file, []byte{0xEF, 0xBB, 0xBF})

	if len(bytes.TrimSpace(file)) == 0 {
		return c.saveDefaults()
	}

	// O campo GameExeFile existiu em versões antigas (o seletor guardava o
	// caminho completo do .exe). Hoje o config guarda só o diretório do jogo;
	// migramos o valor legado e regravamos o JSON sem ele.
	var raw struct {
		*Config
		GameExeFile string `json:"GameExeFile"`
	}
	raw.Config = c
	if err := json.Unmarshal(file, &raw); err != nil {
		return err
	}

	legacyFile := strings.TrimSpace(raw.GameExeFile)
	if legacyFile != "" && strings.TrimSpace(c.GameExeDir) == "" {
		if dir := filepath.Dir(legacyFile); dir != "." && dir != "" {
			c.GameExeDir = dir
			common.LogInfo("GameExeDir derivado do GameExeFile legado: %s", dir)
		}
	}

	return c.validateAndSave(legacyFile != "")
}

// Save salva a configuração no arquivo JSON.
func (c *Config) Save() error {
	return c.save()
}

func (c *Config) saveDefaults() error {
	// Sem escolhas salvas: só o diretório do jogo fica vazio até o usuário
	// (ou o serviço do Steam) informar. O diretório de recursos é derivado.
	c.GameExeDir = ""

	return c.save()
}

func (c *Config) validateAndSave(previouslyChanged bool) error {
	changed := previouslyChanged

	// GameExeDir não é apagado no boot: o drive do jogo pode estar offline
	// (ou a pasta ainda não existir) e perder a configuração salva seria pior
	// do que conviver com um caminho inválido — a validação roda ao abrir o
	// diálogo, que avisa o usuário e deixa corrigir.
	if dir := strings.TrimSpace(c.GameExeDir); dir != "" && !common.IsPathExists(dir) {
		common.LogWarning("GameExeDir não existe: %s", dir)
	}

	// Theme: vazio/"system" segue o sistema; light/dark são escolhas.
	// Valor desconhecido volta para o padrão (sistema).
	theme := strings.ToLower(strings.TrimSpace(c.Theme))
	switch theme {
	case "", "system", "light", "dark":
		if theme != c.Theme {
			c.Theme = theme
			changed = true
		}
	default:
		common.LogWarning("Theme inválido descartado: %q", c.Theme)
		c.Theme = ""
		changed = true
	}

	if changed {
		return c.save()
	}

	return nil
}

func (c *Config) save() error {
	if err := common.EnsurePathExists(c.filePath); err != nil {
		return err
	}

	file, err := os.Create(c.filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			common.LogVerbose("Error closing config file: %v", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}

// Snapshot devolve o espelho da configuração usado pelo frontend.
func (c *Config) Snapshot() *ConfigSnapshot {
	return &ConfigSnapshot{
		ResourcesDir:   c.GetResourcesDir(),
		GameExeDir:     c.GameExeDir,
		PackageBinPath: c.GetPackageBinPath(),
		Theme:          c.ThemePreference(),
	}
}

// GetResourcesDir retorna o diretório de recursos. Ele é FIXO e derivado:
// `resources` ao lado do executável — não é configurável, nem persistido.
func (c *Config) GetResourcesDir() string {
	return filepath.Join(common.GetExecDir(), "resources")
}

// EnsureResourcesDir cria o diretório fixo de recursos. Se não for possível
// criá-lo, o app não tem onde gravar o package.bin: o erro sobe para a UI.
func (c *Config) EnsureResourcesDir() error {
	dir := c.GetResourcesDir()
	if err := common.EnsurePathExists(dir); err != nil {
		return fmt.Errorf("não foi possível preparar o diretório de recursos %s: %w", dir, err)
	}
	return nil
}

// GetGameExeDir retorna o diretório do jogo.
func (c *Config) GetGameExeDir() string {
	return c.GameExeDir
}

// SetGameExeDir define e persiste o diretório do jogo.
func (c *Config) SetGameExeDir(path string) error {
	path = strings.TrimSpace(path)
	if path != "" && !common.IsPathExists(path) {
		return fmt.Errorf("diretório não existe: %s", path)
	}
	c.GameExeDir = path
	return c.save()
}

// ThemePreference devolve a preferência efetiva: "system" quando o
// config.json não tem escolha salva, senão "light" ou "dark".
func (c *Config) ThemePreference() string {
	switch c.Theme {
	case "light", "dark":
		return c.Theme
	default:
		return "system"
	}
}

// GetTheme devolve a preferência efetiva de tema.
func (c *Config) GetTheme() string {
	return c.ThemePreference()
}

// SetTheme persiste a preferência de tema: "system", "light" ou "dark".
// Valor fora desses é recusado (mantém o anterior).
func (c *Config) SetTheme(theme string) error {
	normalized := strings.ToLower(strings.TrimSpace(theme))
	switch normalized {
	case "system", "light", "dark":
	default:
		return fmt.Errorf("tema inválido: use system, light ou dark")
	}
	c.Theme = normalized
	return c.save()
}

// GetPackageBinPath retorna o caminho completo do package.bin (derivado do
// diretório fixo de recursos).
func (c *Config) GetPackageBinPath() string {
	return filepath.Join(c.GetResourcesDir(), "package.bin")
}
