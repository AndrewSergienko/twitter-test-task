package main

import (
	"log/slog"
	"twitter-test-task/internal"
)

func main() {
	app, worker := internal.NewWebApp()

	go worker.RunWorker()
	go worker.RunObserver()

	slog.Error(app.Listen(":3000").Error())
}
