package main

import (
	"log/slog"
	"twitter-test-task/internal"
)

func main() {
	app := internal.NewWebApp()
	slog.Error(app.Listen(":3000").Error())
}
