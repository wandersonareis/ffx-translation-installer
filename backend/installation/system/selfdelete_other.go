//go:build !windows

package system

import "os"

// ScheduleSelfDelete remove o desinstalador. Fora do Windows um executável em
// execução pode ser removido diretamente (unlink).
func ScheduleSelfDelete(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}
