//go:build !windows

package system

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmUninstall pergunta ao usuário (console) se ele quer remover.
func ConfirmUninstall() bool {
	fmt.Print("Remover a tradução instalada? [s/N]: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "s", "sim", "y", "yes":
		return true
	default:
		return false
	}
}

// ShowMessage imprime a mensagem no console.
func ShowMessage(kind MessageKind, title, text string) {
	writer := os.Stdout
	if kind == MessageError {
		writer = os.Stderr
	}
	fmt.Fprintf(writer, "%s\n\n%s\n", title, text)
}
