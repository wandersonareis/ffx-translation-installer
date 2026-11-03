package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// samePath compara dois caminhos absolutos respeitando a case-sensitivity do
// sistema operacional.
func samePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	ca, errA := filepath.Abs(a)
	cb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(ca, cb)
	}
	return ca == cb
}

// isInside informa se target está dentro de base (inclusive igual a base).
func isInside(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// ensureParentDirs cria os diretórios necessários para target e devolve os que
// foram criados, relativos a base e com separador "/".
func ensureParentDirs(base, target string) ([]string, error) {
	parent := filepath.Dir(target)
	if !isInside(base, parent) {
		return nil, fmt.Errorf("caminho fora da pasta do jogo: %s", target)
	}

	rel, err := filepath.Rel(base, parent)
	if err != nil {
		return nil, err
	}
	if rel == "." {
		return nil, nil
	}

	var created []string
	current := base
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Stat(current)
		if err == nil {
			if !info.IsDir() {
				return created, fmt.Errorf("o caminho %s existe e não é uma pasta", current)
			}
			continue
		}
		if !os.IsNotExist(err) {
			return created, err
		}
		if err := os.Mkdir(current, 0o755); err != nil {
			if os.IsExist(err) {
				continue
			}
			return created, err
		}
		relCurrent, err := filepath.Rel(base, current)
		if err != nil {
			return created, err
		}
		created = append(created, filepath.ToSlash(relCurrent))
	}
	return created, nil
}

// removeEmptyDirs remove, do mais profundo para o mais raso, apenas os
// diretórios vazios. Diretórios com arquivos do usuário são preservados.
func removeEmptyDirs(base string, dirs []string) []string {
	unique := make(map[string]struct{}, len(dirs))
	var list []string
	for _, dir := range dirs {
		rel, err := NormalizeRelPath(dir)
		if err != nil {
			continue
		}
		abs := filepath.Join(base, filepath.FromSlash(rel))
		if !isInside(base, abs) || samePath(abs, base) {
			continue
		}
		if _, ok := unique[abs]; ok {
			continue
		}
		unique[abs] = struct{}{}
		list = append(list, abs)
	}

	sort.Slice(list, func(i, j int) bool {
		return pathDepth(list[i]) > pathDepth(list[j])
	})

	var removed []string
	for _, dir := range list {
		if err := os.Remove(dir); err == nil {
			rel, relErr := filepath.Rel(base, dir)
			if relErr == nil {
				removed = append(removed, filepath.ToSlash(rel))
			}
		}
	}
	return removed
}

func pathDepth(path string) int {
	return strings.Count(filepath.Clean(path), string(os.PathSeparator))
}

// copyFileAtomic copia src para dst através de um temporário no mesmo
// diretório de dst, para nunca deixar um desinstalador pela metade.
func copyFileAtomic(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".ffx-uninstaller-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}
