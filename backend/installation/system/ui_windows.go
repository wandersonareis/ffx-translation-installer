//go:build windows

package system

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	messageBoxOK          = 0x00000000
	messageBoxYesNo       = 0x00000004
	messageBoxIconError   = 0x00000010
	messageBoxIconQuery   = 0x00000020
	messageBoxIconWarning = 0x00000030
	messageBoxIconInfo    = 0x00000040
	messageBoxDefButton2  = 0x00000100
	messageBoxIDYes       = 6
)

var (
	user32DLL       = windows.NewLazySystemDLL("user32.dll")
	procMessageBoxW = user32DLL.NewProc("MessageBoxW")
)

// ConfirmUninstall pergunta ao usuário se ele quer remover a tradução.
func ConfirmUninstall() bool {
	answer := messageBox(
		"Desinstalar tradução",
		"Remover a tradução instalada?\n\n"+
			"Somente os arquivos registrados pela tradução serão removidos; "+
			"outros arquivos em data\\mods serão mantidos.",
		messageBoxYesNo|messageBoxIconQuery|messageBoxDefButton2,
	)
	return answer == messageBoxIDYes
}

// ShowMessage exibe uma mensagem nativa do Windows.
func ShowMessage(kind MessageKind, title, text string) {
	icon := uint32(messageBoxIconInfo)
	switch kind {
	case MessageWarning:
		icon = messageBoxIconWarning
	case MessageError:
		icon = messageBoxIconError
	}
	messageBox(title, text, messageBoxOK|icon)
}

func messageBox(title, text string, flags uint32) int {
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return 0
	}
	textPtr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return 0
	}
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)
	return int(ret)
}
