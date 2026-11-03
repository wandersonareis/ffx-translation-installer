package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"ffxinstaller/backend/installation/system"
)

type fakeSource struct {
	info    PackageInfo
	entries []SourceEntry
	data    map[string][]byte
}

func newFakeSource(files map[string]string, modTime time.Time) *fakeSource {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]SourceEntry, 0, len(names))
	data := make(map[string][]byte, len(names))
	for _, name := range names {
		content := []byte(files[name])
		entries = append(entries, SourceEntry{Name: name, Size: int64(len(content)), ModTime: modTime})
		data[name] = content
	}

	return &fakeSource{
		info: PackageInfo{
			Path:          "package.bin",
			Name:          "package.bin",
			Size:          123,
			SHA256:        "abc123",
			Magic:         "FFXX2HDR",
			FormatVersion: 2,
			FileCount:     uint32(len(files)),
			ModifiedAt:    modTime,
		},
		entries: entries,
		data:    data,
	}
}

func (f *fakeSource) Info() PackageInfo { return f.info }

func (f *fakeSource) Entries() []SourceEntry { return f.entries }

func (f *fakeSource) ReadFile(name string) ([]byte, error) {
	content, ok := f.data[name]
	if !ok {
		return nil, fmt.Errorf("arquivo %q não existe no pacote", name)
	}
	return content, nil
}

type fakeRegistry struct {
	entries  []system.UninstallEntry
	removes  int
	failNext bool
}

func (f *fakeRegistry) Register(entry system.UninstallEntry) error {
	if f.failNext {
		return fmt.Errorf("falha simulada no registro")
	}
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeRegistry) Remove() error {
	f.removes++
	return nil
}

func writeFakeExe(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "ffxinstaller-test.exe")
	if err := os.WriteFile(path, []byte("fake-installer-binary"), 0o755); err != nil {
		t.Fatalf("criando exe falso: %v", err)
	}
	return path
}

// tempDir evita a limpeza instável do t.TempDir no Windows: arquivos .exe
// recém-criados podem ficar bloqueados por um instante (antivírus/indexador).
// A limpeza própria roda antes da automática e tenta novamente algumas vezes.
func tempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Cleanup(func() {
		for attempt := 0; attempt < 20; attempt++ {
			if err := os.RemoveAll(dir); err == nil {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	})
	return dir
}
