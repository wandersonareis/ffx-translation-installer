//go:build windows

package common

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

// Um arquivo aberto sem compartilhamento de exclusão (como o jogo pode manter)
// não pode ser apagado. A desinstalação não falha: reporta o arquivo preso e
// mantém o registro/desinstalador para uma nova tentativa.
func TestUninstallReportsLockedFile(t *testing.T) {
	gameDir := tempDir(t)
	registry := installForUninstallTest(t, gameDir)

	target := filepath.Join(gameDir, "data", "mods", "a.txt")
	namePtr, err := windows.UTF16PtrFromString(target)
	if err != nil {
		t.Fatalf("convertendo caminho: %v", err)
	}
	handle, err := windows.CreateFile(
		namePtr,
		windows.GENERIC_READ,
		0, // sem FILE_SHARE_DELETE: simula o arquivo preso
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		t.Fatalf("abrindo arquivo travado: %v", err)
	}
	t.Cleanup(func() { _ = windows.CloseHandle(handle) })

	uninstaller := &Uninstaller{Registry: registry, RunningExe: "/outro/exe"}
	result, err := uninstaller.Uninstall(gameDir)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.ManifestFound {
		t.Fatal("ManifestFound deveria ser true")
	}
	if result.FailedFiles != 1 {
		t.Fatalf("FailedFiles = %d, esperado 1", result.FailedFiles)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("arquivo travado deveria permanecer: %v", err)
	}
	if _, err := os.Stat(ManifestPath(gameDir)); err != nil {
		t.Fatalf("registro deveria permanecer para nova tentativa: %v", err)
	}
	if _, err := os.Stat(UninstallerPath(gameDir)); err != nil {
		t.Fatalf("desinstalador deveria permanecer: %v", err)
	}
	if registry.removes != 0 {
		t.Fatalf("entrada do registro não deveria ser removida: %d", registry.removes)
	}
}
