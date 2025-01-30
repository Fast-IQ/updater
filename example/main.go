package main

import (
	"github.com/Fast-IQ/updater"
	"log/slog"
)

func main() {
	url := "http://127.0.0.1:8080"
	update := updater.NewUpdater()
	err := update.UpdateFile(url, "example/test.txt")
	if err != nil {
		slog.Error("Error update", slog.String("err:", err.Error()))
	}
}
