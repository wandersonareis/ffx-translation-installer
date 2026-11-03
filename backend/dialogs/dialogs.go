// Package dialogs fornece diálogos nativos de arquivo/pasta usando a API do Wails v3.
package dialogs

import (
	"path/filepath"
	"strings"

	"ffxinstaller/backend/common"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// getApp retorna a instância da aplicação Wails.
func getApp() *application.App {
	return application.Get()
}

// SelectGameDir abre o seletor nativo exigindo um dos executáveis do jogo
// (filtro LITERAL: FFX.exe ou FFX-2.exe) e devolve APENAS o diretório que o
// contém. Assim o valor salvo é sempre uma pasta — o mesmo formato do caminho
// devolvido pelo serviço que encontra o jogo no Steam.
func SelectGameDir() string {
	a := getApp()
	if a == nil {
		return ""
	}
	path, err := a.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Selecionar a pasta do jogo (escolha FFX.exe ou FFX-2.exe)",
		Filters: []application.FileFilter{
			{
				DisplayName: "Executável do jogo (FFX.exe; FFX-2.exe)",
				Pattern:     "FFX.exe;FFX-2.exe",
			},
		},
	}).PromptForSingleSelection()
	if err != nil || strings.TrimSpace(path) == "" {
		return ""
	}

	normalized, ok := common.NormalizeGameExe(path)
	if !ok {
		return ""
	}

	return filepath.Dir(normalized)
}
