// Package services fornece os serviços principais da aplicação.
package services

import (
	"ffxinstaller/backend/appwindows"
	"ffxinstaller/backend/config"
	"ffxinstaller/backend/dialogs"
	"ffxinstaller/backend/loggingService"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// configChangedEvent é o eco da configuração: o backend persiste e publica o
// snapshot completo para TODAS as janelas (principal, modal e janela filha),
// que mantêm seus próprios stores em dia sem re-fetch ao reabrir.
const configChangedEvent = "ConfigChanged"

// ConfigService gerencia a configuração do instalador.
type ConfigService struct {
	config *config.Config
}

// OpenWindow abre uma janela filha (usado no desktop para o modal de config).
func (s *ConfigService) OpenWindow(name string) error {
	return appwindows.Open(name)
}

// NewConfigService cria um novo serviço de configuração.
func NewConfigService() *ConfigService {
	return NewConfigServiceWith(config.NewConfig())
}

// NewConfigServiceWith cria o serviço de configuração sobre uma instância já
// carregada. O TranslationService compartilha a mesma instância para não ler
// uma cópia desatualizada do diretório do jogo.
func NewConfigServiceWith(cfg *config.Config) *ConfigService {
	loggingService.Init()
	if cfg == nil {
		cfg = config.NewConfig()
	}
	return &ConfigService{
		config: cfg,
	}
}

// GetConfig retorna a configuração atual em um snapshot com os caminhos
// derivados (ResourcesDir/PackageBinPath) que o frontend exibe.
//
// O diretório de recursos é fixo (<execDir>/resources) e não é configurável:
// se não for possível criá-lo, GetConfig falha e a UI mostra o erro.
func (s *ConfigService) GetConfig() (*config.ConfigSnapshot, error) {
	if err := s.config.EnsureResourcesDir(); err != nil {
		return nil, err
	}
	return s.config.Snapshot(), nil
}

// GetResourcesDir retorna o diretório de recursos (fixo e somente leitura).
func (s *ConfigService) GetResourcesDir() string {
	return s.config.GetResourcesDir()
}

// GetGameExeDir retorna o diretório do jogo.
func (s *ConfigService) GetGameExeDir() string {
	return s.config.GetGameExeDir()
}

// SetGameExeDir define o diretório do jogo.
func (s *ConfigService) SetGameExeDir(path string) error {
	if err := s.config.SetGameExeDir(path); err != nil {
		return err
	}
	s.emitConfigChanged()
	return nil
}

// GetPackageBinPath retorna o caminho do package.bin.
func (s *ConfigService) GetPackageBinPath() string {
	return s.config.GetPackageBinPath()
}

// GetTheme retorna a preferência de tema efetiva ("system"|"light"|"dark").
func (s *ConfigService) GetTheme() string {
	return s.config.GetTheme()
}

// SetTheme persiste a preferência de tema (system, light ou dark) e publica
// o eco para todas as janelas.
func (s *ConfigService) SetTheme(theme string) error {
	if err := s.config.SetTheme(theme); err != nil {
		return err
	}
	s.emitConfigChanged()
	return nil
}

// ValidateGameDir confere a pasta do jogo (data/FFX_Data.vbf e
// data/FFX2_Data.vbf: existência e tamanho aproximado). É só laudo: não
// altera a configuração.
func (s *ConfigService) ValidateGameDir(path string) *config.GameDirValidation {
	return config.ValidateGameDir(path)
}

// SelectGameDir abre o diálogo que exige FFX.exe/FFX-2.exe e devolve apenas a
// pasta do jogo (o executável em si não é persistido).
func (s *ConfigService) SelectGameDir() string {
	return dialogs.SelectGameDir()
}

// emitConfigChanged publica o snapshot pós-mudança no bus do Wails.
//
// O evento foi registrado com o TIPO VALOR (config.ConfigSnapshot): emitir o
// ponteiro faz o Wails recusar o payload ("does not match registered data
// type"), então a snapshot é desreferenciada aqui.
func (s *ConfigService) emitConfigChanged() {
	if app := application.Get(); app != nil {
		app.Event.Emit(configChangedEvent, *s.config.Snapshot())
	}
}
