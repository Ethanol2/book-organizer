package logger

// This package is largely written by gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Log struct {
	Time  string  `json:"time"`
	Level string  `json:"level"`
	Msg   string  `json:"msg"`
	Error *string `json:"error,omitempty"`
}

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

func InitLogger(logsPath string) {
	logDir := filepath.Dir(logsPath)
	if info, err := os.Stat(logsPath); err != nil {
		if err = os.Mkdir(logDir, 0755); err != nil {
			panic("failed to create log directory: " + err.Error())
		}
	} else if !info.IsDir() {
		panic("log directory already exists and is not a directory")
	}

	fileWriter := &lumberjack.Logger{
		Filename:   path.Join(logsPath, "logs.txt"),
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

func GetLogs(filePath string, count, offset int) ([]Log, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, _ := file.Stat()
	filesize := stat.Size()

	var cursor int64 = 0
	var lines [][]byte
	var line []byte

	lineCount := 0

	// Start from the end of the file and move backward
	for int64(len(lines)) < int64(count) && filesize+cursor > 0 {
		cursor--
		file.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		file.Read(char)

		if char[0] == '\n' {
			lineCount++
			if len(line) > 0 {
				// Reverse the line bytes before adding to slice
				lines = append([][]byte{reverse(line)}, lines...)
				line = []byte{}
			}
		} else if lineCount > offset {
			line = append(line, char[0])
		}
	}

	// Add the very first line if we reached start of file
	if len(line) > 0 {
		lines = append([][]byte{reverse(line)}, lines...)
	}

	logs := []Log{}
	errStr := "Failed to parse log"
	for _, line := range lines {
		var newLog Log
		err = json.Unmarshal([]byte(line), &newLog)
		if err != nil {
			newLog = Log{Error: &errStr}
			log.Println(err)
		}
		logs = append(logs, newLog)
	}

	return logs, nil
}

func reverse(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}
