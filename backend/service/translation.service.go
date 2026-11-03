// Package service reúne os serviços de domínio do instalador expostos ao
// frontend via Wails (application.NewService).
package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"ffxinstaller/backend/config"
	installation "ffxinstaller/backend/installation/common"
	"ffxinstaller/backend/loggingService"
	ffxpackage "ffxinstaller/backend/package"
)

// TranslationService instala, consulta e desinstala a tradução.
//
// O registro binário fica na pasta do jogo (ao lado do desinstalador) e é a
// fonte da verdade; no Windows uma entrada em "Aplicativos e recursos" aponta
// para o desinstalador copiado.
type TranslationService struct {
	config *config.Config
}

// NewTranslationService cria o serviço usando a configuração compartilhada.
func NewTranslationService(cfg *config.Config) *TranslationService {
	loggingService.Init()
	if cfg == nil {
		cfg = config.NewConfig()
	}
	return &TranslationService{config: cfg}
}

// Install instala o pacote padrão (resources/package.bin) na pasta do jogo.
func (s *TranslationService) Install() (*installation.Manifest, error) {
	return s.InstallFromPath(s.config.GetPackageBinPath())
}

// InstallFromPath instala um pacote específico na pasta do jogo configurada.
func (s *TranslationService) InstallFromPath(packagePath string) (*installation.Manifest, error) {
	gameDir := s.gameDir()
	if gameDir == "" {
		return nil, fmt.Errorf("diretório do jogo não configurado")
	}
	if !(&GameService{}).IsValid(gameDir) {
		return nil, fmt.Errorf(
			"pasta do jogo inválida: escolha a pasta que contém data\\FFX_Data.vbf e data\\FFX2_Data.vbf",
		)
	}

	pkg, err := ffxpackage.Open(packagePath)
	if err != nil {
		return nil, fmt.Errorf("abrindo pacote %s: %w", packagePath, err)
	}
	defer pkg.Close()

	source, err := installation.NewPackageSource(pkg, packagePath)
	if err != nil {
		return nil, err
	}

	result, err := installation.NewInstaller().Install(source, gameDir)
	if err != nil {
		return nil, err
	}
	for _, warning := range result.Warnings {
		loggingService.Warn("tradução: %s", warning)
	}
	return result.Manifest, nil
}

// Uninstall remove a instalação registrada na pasta do jogo.
func (s *TranslationService) Uninstall() (*installation.UninstallResult, error) {
	gameDir := s.gameDir()
	if gameDir == "" {
		return nil, fmt.Errorf("diretório do jogo não configurado")
	}
	return installation.NewUninstaller().Uninstall(gameDir)
}

// GetInstallation devolve o registro atual ou nil quando não há instalação.
func (s *TranslationService) GetInstallation() (*installation.Manifest, error) {
	gameDir := s.gameDir()
	if gameDir == "" {
		return nil, nil
	}

	manifest, err := installation.LoadManifest(installation.ManifestPath(gameDir))
	if errors.Is(err, installation.ErrManifestNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// IsInstalled informa se existe uma instalação concluída na pasta do jogo.
func (s *TranslationService) IsInstalled() bool {
	manifest, err := s.GetInstallation()
	return err == nil && manifest != nil && manifest.Status == installation.StatusInstalled
}

func (s *TranslationService) gameDir() string {
	dir := strings.TrimSpace(s.config.GetGameExeDir())
	if dir == "" {
		return ""
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}
