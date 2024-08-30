package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

func NewWebApp() *fiber.App {
	app := fiber.New()

	connStr := "postgresql://user1:password@localhost:26257/database?sslmode=disable"
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}

	//messageAdapter := NewMessageAdapter(db)
	messageHandlers := NewMessageHandlers(db, NewEventManager())

	app.Post("/api/messages/", messageHandlers.CreateMessage)
	app.Get("/api/messages/", messageHandlers.GetMessages)

	return app
}
