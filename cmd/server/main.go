package main

import (
	"log/slog"
	"os"

	"github.com/venexene/nango/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		slog.Error("failed to run app", "error", err)
		os.Exit(1)
	}
}
