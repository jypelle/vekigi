package tool

import (
	"log/slog"
	"os"
)

func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
