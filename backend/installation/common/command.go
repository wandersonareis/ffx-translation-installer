package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ffxinstaller/backend/installation/system"
)

const (
	// UninstallFlag é a flag que transforma a cópia do instalador em
	// desinstalador.
	UninstallFlag = "--uninstall"
	quietFlag     = "--yes"
)

// HandleUninstallCommand intercepta "--uninstall" antes do Wails iniciar.
//
// O desinstalador é uma cópia do próprio instalador na pasta do jogo; ele lê o
// registro binário que está ao seu lado. Toda a parte visual e de registro do
// sistema fica no pacote system, selecionada por build tags.
func HandleUninstallCommand(args []string) bool {
	quiet, ok := uninstallArgs(args)
	if !ok {
		return false
	}

	exe, err := os.Executable()
	if err != nil {
		system.ShowMessage(system.MessageError, "Desinstalar tradução",
			"Não foi possível localizar o desinstalador.")
		return true
	}

	// Segurança: só a cópia com o nome canônico pode se auto-apagar. Assim o
	// instalador principal não se remove se alguém rodar "ffxinstaller --uninstall".
	if !strings.EqualFold(filepath.Base(exe), UninstallerFileName) {
		system.ShowMessage(system.MessageError, "Desinstalar tradução",
			"Este executável não é o desinstalador da tradução.")
		return true
	}
	gameDir := filepath.Dir(exe)

	// Sem registro, não há o que confirmar: mostra direto o aviso de remoção
	// manual e apaga o desinstalador.
	if _, statErr := os.Stat(ManifestPath(gameDir)); statErr == nil {
		if !quiet && !system.ConfirmUninstall() {
			return true
		}
	}

	uninstaller := NewUninstaller()
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		system.ShowMessage(system.MessageError, "Desinstalar tradução",
			"Falha ao desinstalar:\n\n"+err.Error())
		return true
	}

	if result.FailedFiles > 0 {
		system.ShowMessage(system.MessageWarning, "Desinstalar tradução",
			"Alguns arquivos não puderam ser removidos.\n\n"+
				"Feche o jogo e execute o desinstalador novamente."+
				detailsText(result))
		return true
	}

	if !result.ManifestFound {
		system.ShowMessage(system.MessageWarning, "Desinstalar tradução",
			"Registro de instalação (.dat) não encontrado.\n\n"+
				"Sem ele não há como saber quais arquivos apagar, então os arquivos "+
				"de data\\mods serão mantidos. Remova-os manualmente se desejar.\n\n"+
				"O desinstalador e a entrada do Windows serão removidos em seguida.")
		if result.SelfDeletePath != "" {
			system.ScheduleSelfDelete(result.SelfDeletePath)
		}
		return true
	}

	system.ShowMessage(system.MessageInfo, "Desinstalar tradução", successText(result))
	// Só agenda a própria exclusão depois que a mensagem foi fechada: no
	// Windows o del falharia com o processo ainda em execução.
	if result.SelfDeletePath != "" {
		system.ScheduleSelfDelete(result.SelfDeletePath)
	}
	return true
}

// uninstallArgs confere se os argumentos pedem a desinstalação.
func uninstallArgs(args []string) (quiet bool, ok bool) {
	if len(args) < 2 || args[1] != UninstallFlag {
		return false, false
	}
	for _, arg := range args[2:] {
		if arg == quietFlag {
			quiet = true
		}
	}
	return quiet, true
}

// successText monta o resumo exibido ao usuário.
func successText(result *UninstallResult) string {
	if result == nil {
		return "Tradução removida."
	}

	var b strings.Builder
	b.WriteString("Tradução removida.\n\n")
	fmt.Fprintf(&b, "Arquivos removidos: %d\n", len(result.RemovedFiles))
	if len(result.RemovedDirs) > 0 {
		fmt.Fprintf(&b, "Pastas removidas: %d\n", len(result.RemovedDirs))
	}
	b.WriteString(detailsText(result))
	return b.String()
}

// detailsText lista os avisos (no máximo 5) para complementar uma mensagem.
func detailsText(result *UninstallResult) string {
	if result == nil || len(result.Warnings) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n\nAvisos:\n")
	for idx, warning := range result.Warnings {
		if idx >= 5 {
			b.WriteString("- ...\n")
			break
		}
		b.WriteString("- " + warning + "\n")
	}
	return b.String()
}
