// Package loggingService é o sistema de logs do aplicativo: console
// colorido por nível + persistência em arquivo (JSON via lumberjack).
// UM arquivo por início de app, em `logs/` ao lado do EXECUTÁVEL.
package loggingService

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once          sync.Once
	log           zerolog.Logger
	consoleLog    zerolog.Logger
	consoleWriter io.Writer = os.Stdout
	logDirOverride string
	activeAppFile *lumberjack.Logger
	logDirPath    string
)

const (
	logLevelEnv    = "LOG_LEVEL"
	logFilePrefix  = "ffxinstaller-"
	logRetention   = 30 * 24 * time.Hour
)

// LogDir retorna o diretório onde os logs são persistidos.
func LogDir() string {
	Init()
	return logDirPath
}

// Get devolve o logger geral (console colorido + arquivo).
func Get() zerolog.Logger {
	Init()
	return log
}

// Init inicializa o sistema de logs uma única vez.
func Init() {
	once.Do(initLoggers)
}

func initLoggers() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	cw := zerolog.ConsoleWriter{
		Out:        consoleOut(),
		TimeFormat: "15:04:05",
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		PartsExclude: []string{zerolog.TimestampFieldName},
	}

	dir := logDir()
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "loggingService: não criou %s: %v\n", dir, err)
		} else {
			pruneOldLogs(dir)
			activeAppFile = rollingFile(filepath.Join(dir, startFileName()))
			logDirPath = dir
		}
	}

	var sink io.Writer = cw
	if activeAppFile != nil {
		sink = zerolog.MultiLevelWriter(cw, activeAppFile)
	}
	log = zerolog.New(sink).
		Level(resolveLevel()).
		With().
		Timestamp().
		Logger()
	consoleLog = zerolog.New(cw).
		Level(resolveLevel()).
		With().
		Timestamp().
		Logger()
}

func closeFiles() {
	if activeAppFile != nil {
		_ = activeAppFile.Close()
	}
	activeAppFile = nil
}

func logDir() string {
	if logDirOverride != "" {
		return logDirOverride
	}
	exe, err := os.Executable()
	if err != nil || exe == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), "logs")
}

func consoleOut() io.Writer {
	if consoleWriter != nil {
		return consoleWriter
	}
	return os.Stdout
}

func rollingFile(name string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   name,
		MaxSize:    5,
		MaxBackups: 10,
		MaxAge:     int(logRetention.Hours() / 24),
		Compress:   true,
	}
}

func resolveLevel() zerolog.Level {
	return zerolog.DebugLevel
}

func startFileName() string {
	return logFilePrefix + time.Now().Format("2006-01-02_15-04-05") + ".log"
}

func pruneOldLogs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	limit := time.Now().Add(-logRetention)
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), logFilePrefix) {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil || !info.ModTime().Before(limit) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}
}

// ResetForTest reinicializa o sistema de logs para testes.
func ResetForTest(dir string) {
	closeFiles()
	once = sync.Once{}
	logDirOverride = dir
	consoleWriter = io.Discard
	logDirPath = ""
	Init()
}

// Façade de log para uso no app

// Debug loga em nível debug.
func Debug(format string, args ...any) {
	log.Debug().Msgf(format, args...)
}

// Info loga em nível info.
func Info(format string, args ...any) {
	log.Info().Msgf(format, args...)
}

// Warn loga em nível warning.
func Warn(format string, args ...any) {
	log.Warn().Msgf(format, args...)
}

// Error loga em nível error.
func Error(format string, args ...any) {
	log.Error().Msgf(format, args...)
}

// Verbose loga se VERBOSE_MODE=1 ou LOG_LEVEL<=debug.
func Verbose(format string, args ...any) {
	if verboseEnabled() || debugEnabled() {
		log.Debug().Msgf(format, args...)
	}
}

func verboseEnabled() bool {
	return os.Getenv("VERBOSE_MODE") == "1"
}

func debugEnabled() bool {
	if verboseEnabled() {
		return true
	}
	if raw := strings.TrimSpace(os.Getenv(logLevelEnv)); raw != "" {
		var level int
		if _, err := fmt.Sscanf(raw, "%d", &level); err == nil && level <= int(zerolog.DebugLevel) {
			return true
		}
	}
	return false
}