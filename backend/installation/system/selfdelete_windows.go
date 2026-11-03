//go:build windows

package system

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// ScheduleSelfDelete apaga o desinstalador depois que ele terminar. O cmd.exe
// espera alguns segundos antes do del, porque o processo ainda está
// encerrando (no Windows não é possível apagar um exe em execução).
func ScheduleSelfDelete(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	script := fmt.Sprintf(`ping 127.0.0.1 -n 3 >nul & del /f /q "%s"`, path)
	cmd := exec.Command("cmd.exe", "/c", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	if err := cmd.Start(); err != nil {
		return
	}
	_ = cmd.Process.Release()
}
