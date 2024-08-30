package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func NewWebApp() (*fiber.App, Worker) {
	app := fiber.New()

	connStr := "postgresql://user1:password@localhost:26257/database?sslmode=disable"
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := amqp.Dial("amqp://user:password@localhost:5672/")

	em := NewEventManager()
	worker := NewWorker(conn, db, em)

	//messageAdapter := NewMessageAdapter(db)
	messageHandlers := NewMessageHandlers(db, em, conn)

	app.Post("/api/messages/", messageHandlers.CreateMessage)
	app.Get("/api/messages/", messageHandlers.GetMessages)

	return app, worker
}
