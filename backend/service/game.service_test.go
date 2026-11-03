package service

import (
	"os"
	"path/filepath"
	"testing"
)

// writeGameData cria os dados obrigatórios dentro da pasta do jogo.
func writeGameData(t *testing.T, dir string) {
	t.Helper()
	for _, rel := range gameDataFiles {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte("dados"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
}

func TestIsValid(t *testing.T) {
	svc := NewGameService()

	if svc.IsValid("") {
		t.Fatal("diretório vazio deveria ser inválido")
	}
	if svc.IsValid("   ") {
		t.Fatal("diretório só com espaços deveria ser inválido")
	}
	if svc.IsValid(filepath.Join(t.TempDir(), "nao-existe")) {
		t.Fatal("diretório inexistente deveria ser inválido")
	}

	// Pasta existe, mas sem data\FFX_Data.vbf e data\FFX2_Data.vbf.
	empty := t.TempDir()
	if svc.IsValid(empty) {
		t.Fatal("pasta sem os dados obrigatórios deveria ser inválida")
	}

	// Com os dois dados obrigatórios, é válida.
	valid := t.TempDir()
	writeGameData(t, valid)
	if !svc.IsValid(valid) {
		t.Fatal("pasta com os dados obrigatórios deveria ser válida")
	}

	// Faltando um dos dados, volta a ser inválida.
	if err := os.Remove(filepath.Join(valid, gameDataFiles[1])); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if svc.IsValid(valid) {
		t.Fatal("faltando um dado obrigatório deveria ser inválido")
	}
}

func TestIsValidRejectsFilePath(t *testing.T) {
	svc := NewGameService()
	file := filepath.Join(t.TempDir(), "FFX.exe")
	if err := os.WriteFile(file, []byte("exe"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if svc.IsValid(file) {
		t.Fatal("caminho de arquivo não é pasta de jogo válida")
	}
}
