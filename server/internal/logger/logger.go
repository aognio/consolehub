package logger

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"consolehub/internal/config"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

type Entry struct {
	Timestamp string         `json:"timestamp"`
	Level     Level          `json:"level"`
	Component string         `json:"component,omitempty"`
	Message   string         `json:"message"`
	Context   map[string]any `json:"context,omitempty"`
}

type Logger struct {
	mu     sync.Mutex
	writer io.Writer
	file   *os.File
	level  Level
}

var globalLogger *Logger
var once sync.Once

func Init(cfg config.LoggingConfig) (*Logger, error) {
	logPath := cfg.ResolvedLogPath()

	dir := filepath.Dir(logPath)
	_ = os.MkdirAll(dir, 0755)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// Fallback to local directory if target path is not writable
		fallbackPath := "./consolehub.log"
		_ = os.MkdirAll(filepath.Dir(fallbackPath), 0755)
		f, _ = os.OpenFile(fallbackPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	}

	var writers []io.Writer
	writers = append(writers, os.Stdout)
	if f != nil {
		writers = append(writers, f)
	}

	l := &Logger{
		writer: io.MultiWriter(writers...),
		file:   f,
		level:  parseLevel(cfg.LogLevel),
	}

	globalLogger = l
	return l, nil
}

func Get() *Logger {
	if globalLogger == nil {
		once.Do(func() {
			_, _ = Init(config.LoggingConfig{LogFile: "./consolehub.log", LogLevel: "debug"})
		})
	}
	return globalLogger
}

func parseLevel(lvl string) Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return LevelDebug
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *Logger) Log(lvl Level, component, message string, ctx map[string]any) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Level:     lvl,
		Component: component,
		Message:   message,
		Context:   ctx,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	data = append(data, '\n')
	_, _ = l.writer.Write(data)
}

func (l *Logger) Debug(component, message string, ctx map[string]any) {
	l.Log(LevelDebug, component, message, ctx)
}

func (l *Logger) Info(component, message string, ctx map[string]any) {
	l.Log(LevelInfo, component, message, ctx)
}

func (l *Logger) Warn(component, message string, ctx map[string]any) {
	l.Log(LevelWarn, component, message, ctx)
}

func (l *Logger) Error(component, message string, ctx map[string]any) {
	l.Log(LevelError, component, message, ctx)
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Package-level helpers
func Debug(component, message string, ctx map[string]any) {
	Get().Debug(component, message, ctx)
}

func Info(component, message string, ctx map[string]any) {
	Get().Info(component, message, ctx)
}

func Warn(component, message string, ctx map[string]any) {
	Get().Warn(component, message, ctx)
}

func Error(component, message string, ctx map[string]any) {
	Get().Error(component, message, ctx)
}
