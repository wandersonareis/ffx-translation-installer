// Package system concentra tudo que depende do sistema operacional:
// registro de desinstalação, confirmação visual e auto-exclusão.
//
// A lógica compartilhada entre sistemas fica em installation/common. Aqui cada
// arquivo é selecionado pelo compilador por meio de build tags.
package system

// UninstallEntry é o que vai para a lista de programas instalados do sistema.
// No Windows, os campos viram valores da chave de desinstalação do registro.
type UninstallEntry struct {
	DisplayName          string
	DisplayVersion       string
	DisplayIcon          string
	InstallLocation      string
	UninstallString      string
	QuietUninstallString string
	InstallDate          string // YYYYMMDD
	EstimatedSizeKB      uint32
}

// UninstallRegistry registra o desinstalador no sistema operacional.
// Linux/macOS usam uma implementação vazia (o registro binário na pasta do
// jogo continua sendo a fonte da verdade).
type UninstallRegistry interface {
	Register(entry UninstallEntry) error
	Remove() error
}

// MessageKind classifica a mensagem exibida ao usuário.
type MessageKind int

const (
	// MessageInfo é uma mensagem informativa.
	MessageInfo MessageKind = iota
	// MessageWarning é um aviso.
	MessageWarning
	// MessageError é um erro.
	MessageError
)
