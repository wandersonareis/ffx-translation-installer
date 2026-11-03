// Package appwindows gerencia as janelas filhas do app (frameless,
// conteúdo HTML/React dentro — o mesmo papel dos modais, mas como janela de
// verdade: arrastável, com botões de janela e foco próprio).
package appwindows

import (
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// MainWindowName é o Name da janela principal (main.go). Toda janela filha
// nasce ligada a ela e morre com ela.
const MainWindowName = "main"

// windowMu serializa Open: duas chamadas simultâneas que não acham a janela
// criariam DUAS janelas com o mesmo Name (a registry é por ID, não por nome).
var windowMu sync.Mutex

// closeChildrenHook garante (uma única vez) que o fecho da principal fecha
// todas as filhas abertas — cobre o caso de o SO destruir a principal sem
// passar pelo fluxo de owner do Windows.
var closeChildrenHook sync.Once

type spec struct {
	title             string
	width, height     int
	minWidth, minHght int
}

var specs = map[string]spec{
	// Altura = titlebar (36px) + o único campo (label/input/hint) e a mensagem
	// de validação em uma linha, mais o padding do corpo (16px em cima/baixo).
	"config": {title: "Configurações", width: 720, height: 188, minWidth: 600, minHght: 188},
}

// Open abre (ou foca, se já aberta) a janela filha pelo nome.
func Open(name string) error {
	s, ok := specs[name]
	if !ok {
		return fmt.Errorf("janela desconhecida: %q", name)
	}
	a := application.Get()
	if a == nil {
		return fmt.Errorf("app não inicializado")
	}
	windowMu.Lock()
	defer windowMu.Unlock()

	if w, found := a.Window.GetByName(name); found {
		w.Show()
		w.Focus()
		return nil
	}

	opts := application.WebviewWindowOptions{
		Name:      name,
		Title:     s.title,
		Width:     s.width,
		Height:    s.height,
		MinWidth:  s.minWidth,
		MinHeight: s.minHght,
		BackgroundColour: application.NewRGB(21, 22, 26),
		Frameless: true,
		// Hash na raiz: o asset server do Wails só serve index.html em "/"
		// (não há fallback de SPA), então "/w/#config" abriria vazio. O
		// window-host do frontend lê o hash e monta o conteúdo certo.
		URL: "/#config",
	}
	w := a.Window.NewWithOptions(opts)
	w.Center()
	attachToMain(a, w)
	return nil
}

// attachToMain liga a filha MODAL à principal: o Windows a mantém acima e
// desabilita a principal enquanto ela estiver aberta.
func attachToMain(a *application.App, w *application.WebviewWindow) {
	parent, ok := a.Window.GetByName(MainWindowName)
	if !ok {
		return
	}
	ensureCloseChildrenHook(a)
	parent.AttachModal(w)
	// Devolução do foco: o Wails só reabilita a owner DEPOIS do DestroyWindow;
	// reabilitar aqui (hooks rodam antes dos listeners) deixa a ativação já
	// na principal quando a filha morre.
	w.RegisterHook(events.Common.WindowClosing, func(*application.WindowEvent) {
		parent.SetEnabled(true)
		parent.Focus()
	})
}

// ensureCloseChildrenHook garante que o fecho da principal fecha todas as
// filhas abertas (sem isso uma filha órfã pode segurar o quit no Windows).
func ensureCloseChildrenHook(a *application.App) {
	parent, ok := a.Window.GetByName(MainWindowName)
	if !ok {
		return
	}
	closeChildrenHook.Do(func() {
		parent.RegisterHook(events.Common.WindowClosing, func(*application.WindowEvent) {
			for name := range specs {
				if child, found := a.Window.GetByName(name); found {
					child.Close()
				}
			}
		})
	})
}
