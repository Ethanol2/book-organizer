package logger

// This package is largely written by gemini

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

const LogsFilePath string = "./data/logs/"

// 1. Define a completely custom Console Handler
type CustomConsoleHandler struct {
	writer io.Writer
}

func (h *CustomConsoleHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *CustomConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler     { return h }
func (h *CustomConsoleHandler) WithGroup(name string) slog.Handler           { return h }

func (h *CustomConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	// Format the time code cleanly (e.g., "13:45:02")
	timeStr := r.Time.Format("15:04:05")

	// Choose terminal colors based on the log level
	var levelStr string
	switch r.Level {
	case slog.LevelDebug:
		levelStr = "\033[35m[DEBUG]\033[0m" // Magenta
	case slog.LevelInfo:
		levelStr = "\033[34m[INFO ]\033[0m" // Blue
	case slog.LevelWarn:
		levelStr = "\033[33m[WARN ]\033[0m" // Yellow
	case slog.LevelError:
		levelStr = "\033[31m[ERROR]\033[0m" // Red
	default:
		levelStr = fmt.Sprintf("[%s]", r.Level.String())
	}

	// Build the main string payload
	output := fmt.Sprintf("%s %s %s", timeStr, levelStr, r.Message)

	// Append any structured key-value variables passed to slog (if any exist)
	r.Attrs(func(a slog.Attr) bool {
		output += fmt.Sprintf(" \033[90m%s=%v\033[0m", a.Key, a.Value.Any())
		return true
	})

	// Write directly to terminal with a newline
	_, err := fmt.Fprintln(h.writer, output)
	return err
}

// 2. Updated DualHandler managing our fully custom layout
type DualHandler struct {
	consoleHandler slog.Handler
	fileHandler    slog.Handler
}

func (h *DualHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *DualHandler) Handle(ctx context.Context, r slog.Record) error {
	_ = h.consoleHandler.Handle(ctx, r.Clone())
	_ = h.fileHandler.Handle(ctx, r.Clone())
	return nil
}

func (h *DualHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *DualHandler) WithGroup(name string) slog.Handler       { return h }

func InitLogger() {
	logDir := filepath.Dir(LogsFilePath)
	if info, err := os.Stat(LogsFilePath); err != nil {
		if err = os.Mkdir(logDir, 0755); err != nil {
			panic("failed to create log directory: " + err.Error())
		}
	} else if !info.IsDir() {
		panic("log directory already exists and is not a directory")
	}

	fileWriter := &lumberjack.Logger{
		Filename:   path.Join(LogsFilePath, "logs.txt"),
		MaxSize:    10,
		MaxBackups: 3,
		Compress:   true,
	}

	// Custom terminal output + standard structured file output
	customHandler := &DualHandler{
		consoleHandler: &CustomConsoleHandler{writer: os.Stdout},
		fileHandler:    slog.NewJSONHandler(fileWriter, nil),
	}

	slog.SetDefault(slog.New(customHandler))
	log.SetFlags(0)
}
