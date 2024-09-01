package internal

import (
	"github.com/gofiber/fiber/v2"
)

func NewWebApp(ioc IoC, em *EventManager) *fiber.App {
	app := fiber.New()

	messageHandlers := NewMessageHandlers(em, ioc)

	app.Post("/api/messages/", messageHandlers.CreateMessage)
	app.Get("/api/messages/", messageHandlers.GetMessages)

	return app
}
