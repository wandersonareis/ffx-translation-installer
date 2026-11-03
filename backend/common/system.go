// Package common contém utilitários compartilhados.
package common

import (
	"fmt"
	"os"
	"path/filepath"

	"ffxinstaller/backend/loggingService"
)

// GetExecDir retorna o diretório do executável.
func GetExecDir() string {
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}

// EnsurePathExists garante que o diretório do caminho existe.
func EnsurePathExists(path string) error {
	cPath := filepath.Clean(path)
	if filepath.Ext(cPath) != "" {
		cPath = filepath.Dir(cPath)
	}
	if err := os.MkdirAll(cPath, os.ModePerm); err != nil {
		return fmt.Errorf("erro ao criar diretório %s: %w", cPath, err)
	}
	return nil
}

// IsPathExists verifica se o caminho existe.
func IsPathExists(path string) bool {
	_, err := os.Stat(filepath.Clean(path))
	return !os.IsNotExist(err)
}

// Façade de log: delega ao loggingService.

func LogVerbose(format string, args ...any) {
	loggingService.Verbose(format, args...)
}

func LogInfo(format string, args ...any) {
	loggingService.Info(format, args...)
}

func LogWarning(format string, args ...any) {
	loggingService.Warn(format, args...)
}

func LogError(format string, args ...any) {
	loggingService.Error(format, args...)
}