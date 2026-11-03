package common

import (
	"fmt"
	"path"
	"strings"
)

// invalidRelPathChars são caracteres que o Windows não aceita em nomes de
// arquivo. Rejeitá-los no pacote evita erros obscuros no meio da instalação.
const invalidRelPathChars = `<>:"|?*`

// NormalizeRelPath valida um nome de arquivo vindo do pacote e devolve o
// caminho relativo normalizado (separador "/").
//
// O pacote nunca pode escrever fora da pasta do jogo: caminhos absolutos,
// unidades (C:), ".." e nomes reservados do Windows são recusados.
func NormalizeRelPath(name string) (string, error) {
	raw := strings.TrimSpace(name)
	if raw == "" {
		return "", fmt.Errorf("caminho vazio no pacote")
	}
	if strings.ContainsRune(raw, 0) {
		return "", fmt.Errorf("caminho com byte nulo: %q", name)
	}

	s := strings.ReplaceAll(raw, `\`, "/")
	if strings.HasPrefix(s, "/") {
		return "", fmt.Errorf("caminho absoluto não permitido: %q", name)
	}
	if strings.HasSuffix(s, "/") {
		return "", fmt.Errorf("caminho de pasta não permitido: %q", name)
	}
	if len(s) >= 2 && s[1] == ':' {
		return "", fmt.Errorf("caminho com unidade não permitido: %q", name)
	}

	for _, r := range s {
		if r < 0x20 || strings.ContainsRune(invalidRelPathChars, r) {
			return "", fmt.Errorf("caminho com caractere inválido: %q", name)
		}
	}

	clean := path.Clean(s)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("caminho fora da pasta do jogo: %q", name)
	}

	for _, part := range strings.Split(clean, "/") {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("caminho inválido: %q", name)
		}
		if isWindowsReservedName(part) {
			return "", fmt.Errorf("nome reservado do Windows: %q", name)
		}
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return "", fmt.Errorf("nome com ponto/espaço no final: %q", name)
		}
	}

	return clean, nil
}

// isWindowsReservedName confere nomes de dispositivo do Windows (CON, NUL,
// COM1 etc.), com ou sem extensão.
func isWindowsReservedName(part string) bool {
	base := part
	if dot := strings.IndexByte(base, '.'); dot >= 0 {
		base = base[:dot]
	}
	base = strings.ToUpper(strings.TrimSpace(base))

	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	if len(base) == 4 {
		if strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT") {
			return base[3] >= '1' && base[3] <= '9'
		}
	}
	return false
}
