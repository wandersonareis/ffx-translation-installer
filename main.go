package main

import (
	"embed"
	"os"

	"ffxinstaller/backend/config"
	installation "ffxinstaller/backend/installation/common"
	"ffxinstaller/backend/loggingService"
	"ffxinstaller/backend/service"
	"ffxinstaller/backend/services"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// Registra o payload do evento de eco da configuração: a mesma snapshot
	// atravessa todas as janelas (principal, modal e janela filha).
	application.RegisterEvent[config.ConfigSnapshot]("ConfigChanged")
}

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts the application loop. Logs vão para arquivo (logs/) e console.
func main() {
	loggingService.Init()

	// A cópia do instalador na pasta do jogo roda com "--uninstall": a
	// desinstalação acontece antes de subir o Wails, com confirmação nativa.
	if installation.HandleUninstallCommand(os.Args) {
		return
	}

	loggingService.Info("ffxinstaller iniciado - log em %s", loggingService.LogDir())

	// Config compartilhada: o ConfigService persiste e o TranslationService
	// enxerga o mesmo diretório do jogo e o mesmo package.bin.
	cfg := config.NewConfig()

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	// ServerOptions enables server mode (-tags server): runs as HTTP server without native GUI.
	app := application.New(application.Options{
		Name:        "ffxinstaller",
		Description: "Installer Translation FINAL FANTASY X and X-2 HD Remaster",
		Services: []application.Service{
			application.NewService(services.NewConfigServiceWith(cfg)),
			application.NewService(service.NewGameService()),
			application.NewService(service.NewTranslationService(cfg)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Server: application.ServerOptions{
			Host: "localhost",
			Port: 9246,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create a new window with the necessary options.
	// 'Name' identifica a principal no registry de appwindows: as janelas
	// filhas nascem ligadas a ela (AttachModal) e fecham junto com ela.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	//
	// Medidas = 70% de 1280x720 (30% menor), mantendo a proporção 16:9.
	// Sempre números PARES: 896x504 (ímpar subiria 1).
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:          "main",
		Title:         "Installer Translation FINAL FANTASY X and X-2 HD Remaster",
		Width:         896,
		Height:        504,
		MinWidth:      896,
		MinHeight:     504,
		MaxWidth:      896,
		MaxHeight:     504,
		DisableResize: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(21, 22, 26),
		URL:              "/",
	})

	// Run the application. This blocks until the application has been exited.
	if err := app.Run(); err != nil {
		loggingService.Error("erro ao executar a aplicação: %v", err)
		os.Exit(1)
	}
}
