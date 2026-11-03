package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSized(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestValidateGameDirValid(t *testing.T) {
	dir := t.TempDir()
	files := []expectedGameFile{
		{RelPath: filepath.Join("data", "FFX_Data.vbf"), Expected: 1000, Tolerance: 0.20},
		{RelPath: filepath.Join("data", "FFX2_Data.vbf"), Expected: 500, Tolerance: 0.20},
	}
	writeSized(t, filepath.Join(dir, "data", "FFX_Data.vbf"), 1000)
	writeSized(t, filepath.Join(dir, "data", "FFX2_Data.vbf"), 450)
	writeSized(t, filepath.Join(dir, "FFX.exe"), 10)
	writeSized(t, filepath.Join(dir, "FFX-2.exe"), 10)

	res := validateGameDir(dir, files, optionalGameExes)
	if !res.Configured {
		t.Fatal("pasta informada deveria estar configurada")
	}
	if !res.Valid {
		t.Fatalf("pasta válida recusada: %v", res.Problems)
	}
	if len(res.Problems) != 0 {
		t.Fatalf("problemas inesperados: %v", res.Problems)
	}
	if len(res.Warnings) != 0 {
		t.Fatalf("avisos inesperados: %v", res.Warnings)
	}
	if len(res.Files) != 2 || !res.Files[0].OK || !res.Files[1].OK {
		t.Fatalf("status dos arquivos inesperado: %+v", res.Files)
	}
}

func TestValidateGameDirMissingFile(t *testing.T) {
	dir := t.TempDir()
	files := []expectedGameFile{
		{RelPath: filepath.Join("data", "FFX_Data.vbf"), Expected: 1000, Tolerance: 0.20},
		{RelPath: filepath.Join("data", "FFX2_Data.vbf"), Expected: 500, Tolerance: 0.20},
	}
	writeSized(t, filepath.Join(dir, "data", "FFX_Data.vbf"), 1000)

	res := validateGameDir(dir, files, nil)
	if res.Valid {
		t.Fatal("faltando FFX2_Data.vbf deveria invalidar")
	}
	if len(res.Problems) != 1 || !strings.Contains(res.Problems[0], "FFX2_Data.vbf") {
		t.Fatalf("problema esperado para FFX2_Data.vbf: %v", res.Problems)
	}
}

func TestValidateGameDirTruncated(t *testing.T) {
	dir := t.TempDir()
	files := []expectedGameFile{
		{RelPath: filepath.Join("data", "FFX_Data.vbf"), Expected: 1000, Tolerance: 0.20},
	}
	// 700 bytes = 30% abaixo da referência de 1000.
	writeSized(t, filepath.Join(dir, "data", "FFX_Data.vbf"), 700)

	res := validateGameDir(dir, files, nil)
	if res.Valid {
		t.Fatal("arquivo truncado deveria invalidar")
	}
	if len(res.Problems) != 1 || !strings.Contains(res.Problems[0], "Tamanho inesperado") {
		t.Fatalf("problema de tamanho esperado: %v", res.Problems)
	}
}

func TestValidateGameDirEmptyAndMissing(t *testing.T) {
	empty := validateGameDir("   ", expectedGameFiles, optionalGameExes)
	if empty.Configured || empty.Valid {
		t.Fatal("pasta vazia não é configuração válida")
	}
	if len(empty.Warnings) == 0 {
		t.Fatal("pasta vazia deveria orientar a configuração")
	}

	missing := validateGameDir(filepath.Join(t.TempDir(), "nao-existe"), expectedGameFiles, optionalGameExes)
	if missing.Valid {
		t.Fatal("pasta inexistente deveria invalidar")
	}
	if len(missing.Problems) != 1 || !strings.Contains(missing.Problems[0], "Pasta não encontrada") {
		t.Fatalf("problema esperado de pasta: %v", missing.Problems)
	}
}

func TestValidateGameDirExeWarningOnly(t *testing.T) {
	dir := t.TempDir()
	files := []expectedGameFile{
		{RelPath: filepath.Join("data", "FFX_Data.vbf"), Expected: 100, Tolerance: 0.20},
	}
	writeSized(t, filepath.Join(dir, "data", "FFX_Data.vbf"), 100)

	res := validateGameDir(dir, files, optionalGameExes)
	if !res.Valid {
		t.Fatalf("falta de .exe não deveria invalidar: %v", res.Problems)
	}
	if len(res.Warnings) != 2 {
		t.Fatalf("esperava aviso para os dois executáveis: %v", res.Warnings)
	}
}
