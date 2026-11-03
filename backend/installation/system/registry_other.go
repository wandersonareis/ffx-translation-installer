//go:build !windows

package system

type noopRegistry struct{}

// NewUninstallRegistry devolve um registro vazio: fora do Windows a lista de
// programas instalados é gerenciada pelo próprio pacote/instalador.
func NewUninstallRegistry(string) UninstallRegistry { return noopRegistry{} }

func (noopRegistry) Register(UninstallEntry) error { return nil }

func (noopRegistry) Remove() error { return nil }
