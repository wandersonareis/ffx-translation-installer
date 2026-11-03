package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// expectedGameFile é um arquivo que o instalador precisa encontrar dentro da
// pasta do jogo. Expected é o tamanho de referência (bytes) medido na
// instalação do FINAL FANTASY X/X-2 HD Remaster; Tolerance aceita variações
// de edição/idioma sem deixar passar download truncado.
type expectedGameFile struct {
	RelPath   string
	Expected  int64
	Tolerance float64
}

// expectedGameFiles são os dados do jogo — mais importantes que os .exe: é
// neles que a tradução será aplicada.
var expectedGameFiles = []expectedGameFile{
	{RelPath: filepath.Join("data", "FFX_Data.vbf"), Expected: 20701622962, Tolerance: 0.20},   // ~19,3 GiB
	{RelPath: filepath.Join("data", "FFX2_Data.vbf"), Expected: 17867755892, Tolerance: 0.20}, // ~16,6 GiB
}

// optionalGameExes não invalidam a pasta (o usuário pode ter outra edição),
// mas a ausência é avisada no diálogo de configuração.
var optionalGameExes = []string{"FFX.exe", "FFX-2.exe"}

// GameFileStatus é o resultado da checagem de um arquivo obrigatório.
type GameFileStatus struct {
	RelPath  string `json:"RelPath"`
	Exists   bool   `json:"Exists"`
	Size     int64  `json:"Size"`
	Expected int64  `json:"Expected"`
	OK       bool   `json:"OK"`
}

// GameDirValidation é o laudo da pasta do jogo consumido pelo frontend.
// Problems invalidam a pasta; Warnings são informativos.
type GameDirValidation struct {
	Path       string           `json:"Path"`
	Configured bool             `json:"Configured"`
	Valid      bool             `json:"Valid"`
	Problems   []string         `json:"Problems"`
	Warnings   []string         `json:"Warnings"`
	Files      []GameFileStatus `json:"Files"`
}

// ValidateGameDir confere a existência e o tamanho aproximado dos dados
// obrigatórios (data/FFX_Data.vbf e data/FFX2_Data.vbf). Nunca altera a
// configuração: é só laudo para a UI.
func ValidateGameDir(path string) *GameDirValidation {
	return validateGameDir(path, expectedGameFiles, optionalGameExes)
}

func validateGameDir(path string, files []expectedGameFile, exes []string) *GameDirValidation {
	res := &GameDirValidation{
		Path:  strings.TrimSpace(path),
		Files: make([]GameFileStatus, 0, len(files)),
	}
	if res.Path == "" {
		res.Warnings = append(res.Warnings, "Configure a pasta onde o jogo está instalado.")
		return res
	}
	res.Configured = true

	info, err := os.Stat(res.Path)
	if err != nil || !info.IsDir() {
		res.Problems = append(res.Problems, fmt.Sprintf("Pasta não encontrada: %s", res.Path))
		return res
	}

	valid := true
	for _, f := range files {
		status := GameFileStatus{RelPath: f.RelPath, Expected: f.Expected}
		full := filepath.Join(res.Path, f.RelPath)
		st, err := os.Stat(full)
		if err != nil {
			res.Problems = append(res.Problems, fmt.Sprintf("Arquivo não encontrado: %s", f.RelPath))
			res.Files = append(res.Files, status)
			valid = false
			continue
		}

		status.Exists = true
		status.Size = st.Size()
		low := int64(float64(f.Expected) * (1 - f.Tolerance))
		high := int64(float64(f.Expected) * (1 + f.Tolerance))
		status.OK = status.Size >= low && status.Size <= high
		if !status.OK {
			res.Problems = append(res.Problems, fmt.Sprintf(
				"Tamanho inesperado em %s: %s (esperado ~%s)",
				f.RelPath, formatBytes(status.Size), formatBytes(f.Expected)))
			valid = false
		}
		res.Files = append(res.Files, status)
	}

	for _, name := range exes {
		if _, err := os.Stat(filepath.Join(res.Path, name)); err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("Executável não encontrado: %s", name))
		}
	}

	res.Valid = valid
	return res
}

// formatBytes apresenta tamanhos grandes em GiB (decimal) para as mensagens.
func formatBytes(n int64) string {
	const gib = 1 << 30
	if n >= gib {
		return fmt.Sprintf("%.1f GiB", float64(n)/gib)
	}
	return fmt.Sprintf("%d bytes", n)
}
