//go:build windows

package system

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type windowsRegistry struct {
	productID string
}

// NewUninstallRegistry devolve o registro de desinstalação do Windows para o
// produto informado. HKCU não exige elevação.
func NewUninstallRegistry(productID string) UninstallRegistry {
	return windowsRegistry{productID: productID}
}

func (r windowsRegistry) keyPath() string {
	return `Software\Microsoft\Windows\CurrentVersion\Uninstall\` + r.productID
}

func (r windowsRegistry) Register(entry UninstallEntry) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, r.keyPath(), registry.WRITE)
	if err != nil {
		return fmt.Errorf("abrindo chave de desinstalação: %w", err)
	}
	defer key.Close()

	values := []struct {
		name  string
		value string
	}{
		{"DisplayName", entry.DisplayName},
		{"DisplayVersion", entry.DisplayVersion},
		{"DisplayIcon", entry.DisplayIcon},
		{"InstallLocation", entry.InstallLocation},
		{"UninstallString", entry.UninstallString},
		{"QuietUninstallString", entry.QuietUninstallString},
		{"InstallDate", entry.InstallDate},
	}
	for _, v := range values {
		if v.value == "" {
			continue
		}
		if err := key.SetStringValue(v.name, v.value); err != nil {
			return fmt.Errorf("gravando %s: %w", v.name, err)
		}
	}

	if err := key.SetDWordValue("EstimatedSize", entry.EstimatedSizeKB); err != nil {
		return fmt.Errorf("gravando EstimatedSize: %w", err)
	}
	if err := key.SetDWordValue("NoModify", 1); err != nil {
		return fmt.Errorf("gravando NoModify: %w", err)
	}
	if err := key.SetDWordValue("NoRepair", 1); err != nil {
		return fmt.Errorf("gravando NoRepair: %w", err)
	}
	return nil
}

func (r windowsRegistry) Remove() error {
	err := registry.DeleteKey(registry.CURRENT_USER, r.keyPath())
	if err == registry.ErrNotExist {
		return nil
	}
	if err != nil {
		return fmt.Errorf("removendo chave de desinstalação: %w", err)
	}
	return nil
}
