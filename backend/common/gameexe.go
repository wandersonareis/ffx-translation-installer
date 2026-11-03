package common

import (
	"path/filepath"
	"strings"
)

// gameExeNames é o PAR de executáveis aceitos. O seletor de arquivo e o
// campo de configuração não podem virar seletor de qualquer .exe do disco:
// só FFX.exe e FFX-2.exe, sem curinga.
var gameExeNames = map[string]bool{
	"ffx.exe":   true,
	"ffx-2.exe": true,
}

// NormalizeGameExe valida e normaliza o caminho do executável do jogo.
//
// "" é aceito (limpar a configuração). Qualquer outro caminho só passa se o
// nome base for FFX.exe ou FFX-2.exe, sem diferenciar maiúsculas — devolve o
// caminho como veio (só com TrimSpace). Caminho rejeitado sai como ("", false)
// e NÃO deve ser persistido.
func NormalizeGameExe(path string) (string, bool) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", true
	}
	if !gameExeNames[strings.ToLower(filepath.Base(p))] {
		return "", false
	}
	return p, true
}
