package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ffxinstaller/backend/installation/system"
)

// UninstallResult descreve o que foi removido pelo desinstalador.
type UninstallResult struct {
	// ManifestFound informa se o registro binário foi encontrado e lido.
	ManifestFound      bool     `json:"ManifestFound"`
	RemovedFiles       []string `json:"RemovedFiles"`
	RemovedDirs        []string `json:"RemovedDirs"`
	Warnings           []string `json:"Warnings,omitempty"`
	RegistryRemoved    bool     `json:"RegistryRemoved"`
	UninstallerRemoved bool     `json:"UninstallerRemoved"`
	// FailedFiles conta arquivos registrados que existiam mas não puderam ser
	// removidos (por exemplo, jogo aberto). Nesse caso o registro e o
	// desinstalador são mantidos para nova tentativa.
	FailedFiles int `json:"FailedFiles"`
	// SelfDeletePath é preenchido quando o desinstalador em execução não pôde
	// se remover (Windows); o chamador agenda a exclusão.
	SelfDeletePath string `json:"SelfDeletePath,omitempty"`
}

// Uninstaller remove a instalação registrada na pasta do jogo.
type Uninstaller struct {
	// Registry é o registro do sistema (injetável nos testes).
	Registry system.UninstallRegistry
	// RunningExe é o executável em execução; quando é o próprio desinstalador,
	// ele é removido depois (SelfDeletePath).
	RunningExe string
}

// NewUninstaller cria um desinstalador usando o registro do sistema e o
// executável atual.
func NewUninstaller() *Uninstaller {
	exe, _ := os.Executable()
	return &Uninstaller{
		Registry:   system.NewUninstallRegistry(ProductID),
		RunningExe: exe,
	}
}

// Uninstall remove somente os arquivos listados no registro. Arquivos do
// usuário em data/mods que não pertencem à tradução são preservados.
//
// A operação é tolerante: registro ou arquivos ausentes são ignorados e
// arquivos presos são pulados com aviso. Só erros de I/O inesperados (por
// exemplo, não conseguir resolver a pasta do jogo) voltam como erro.
func (u *Uninstaller) Uninstall(gameDir string) (*UninstallResult, error) {
	gameDir = strings.TrimSpace(gameDir)
	if gameDir == "" {
		return nil, fmt.Errorf("diretório do jogo não informado")
	}
	absGameDir, err := filepath.Abs(gameDir)
	if err != nil {
		return nil, fmt.Errorf("resolvendo diretório do jogo: %w", err)
	}

	result := &UninstallResult{}

	manifest, err := LoadManifest(ManifestPath(absGameDir))
	switch {
	case errors.Is(err, ErrManifestNotFound), errors.Is(err, ErrManifestCorrupt):
		// Sem o .dat não há lista de arquivos para apagar. Removemos só o que
		// não depende dele: o .dat inválido, o desinstalador e a entrada do
		// sistema. Os arquivos em data/mods são preservados.
		result.Warnings = append(result.Warnings,
			"registro de instalação não encontrado; os arquivos não serão removidos (remova data/mods manualmente se desejar)")
		u.cleanUp(absGameDir, result)
		return result, nil
	case err != nil:
		return nil, err
	}

	result.ManifestFound = true
	failedFiles := 0

	for _, file := range manifest.Files {
		rel, err := NormalizeRelPath(file.Path)
		if err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("caminho inválido ignorado no registro: %s", file.Path))
			continue
		}

		target := filepath.Join(absGameDir, filepath.FromSlash(rel))
		if !isInside(absGameDir, target) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("caminho fora da pasta do jogo ignorado: %s", file.Path))
			continue
		}

		info, err := os.Lstat(target)
		if err != nil {
			if os.IsNotExist(err) {
				// Arquivo já removido: nada a fazer.
				continue
			}
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("não foi possível verificar %s: %v", rel, err))
			continue
		}
		if info.IsDir() {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("ignorando pasta registrada como arquivo: %s", rel))
			continue
		}

		if err := os.Remove(target); err != nil {
			failedFiles++
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("não foi possível remover %s: %v", rel, err))
			continue
		}
		result.RemovedFiles = append(result.RemovedFiles, rel)
	}

	result.RemovedDirs = removeEmptyDirs(absGameDir, manifest.CreatedDirs)

	// Arquivo preso (jogo aberto, antivírus etc.): mantém registro,
	// desinstalador e entrada do sistema para uma nova tentativa.
	if failedFiles > 0 {
		result.FailedFiles = failedFiles
		return result, nil
	}

	u.cleanUp(absGameDir, result)
	return result, nil
}

// cleanUp remove o registro binário, o desinstalador e a entrada do sistema.
// Arquivos inexistentes são ignorados.
func (u *Uninstaller) cleanUp(absGameDir string, result *UninstallResult) {
	if err := os.Remove(ManifestPath(absGameDir)); err != nil && !os.IsNotExist(err) {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("não foi possível remover o registro de instalação: %v", err))
	}

	uninstallerPath := UninstallerPath(absGameDir)
	switch {
	case u.RunningExe != "" && samePath(uninstallerPath, u.RunningExe):
		// No Windows um executável em execução não pode ser apagado; o
		// chamador agenda a remoção.
		result.SelfDeletePath = uninstallerPath
	default:
		if err := os.Remove(uninstallerPath); err != nil {
			if os.IsNotExist(err) {
				result.UninstallerRemoved = true
			} else {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("não foi possível remover o desinstalador: %v", err))
			}
		} else {
			result.UninstallerRemoved = true
		}
	}

	if err := u.Registry.Remove(); err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("não foi possível remover a entrada do registro do sistema: %v", err))
	} else {
		result.RegistryRemoved = true
	}
}
