package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valyala/fasthttp"
	"log/slog"
)

type MessageSaver interface {
	SaveMessage(message Message) error
}

type MessageReader interface {
	GetMessages() ([]Message, error)
}

//type MessageHandlers struct {
//	messageSaver  MessageSaver
//	messageReader MessageReader
//}

type MessageHandlers struct {
	eventManager *EventManager
	db           *sqlx.DB
	conn         *amqp.Connection
}

// func NewMessageHandlers(ms MessageSaver, mr MessageReader) *MessageHandlers {
func NewMessageHandlers(db *sqlx.DB, em *EventManager, conn *amqp.Connection) *MessageHandlers {
	return &MessageHandlers{
		db:           db,
		eventManager: em,
		conn:         conn,
	}
}

func (container MessageHandlers) CreateMessage(c *fiber.Ctx) error {
	var requestData struct {
		Nickname string `json:"nickname"`
		Text     string `json:"text"`
	}

	if err := c.BodyParser(&requestData); err != nil {
		slog.Warn(fmt.Sprintf("Cannot parse JSON: %v", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	tx := container.db.MustBegin()
	messageSaver := NewMessageAdapter(tx, container.conn)

	message := Message{
		Nickname: requestData.Nickname,
		Text:     requestData.Text,
	}

	if messageSaver.RequestSaveMessage(message) != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot save message"})
	}
	err := tx.Commit()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok"})
}

func (container MessageHandlers) GetMessages(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	//tx := container.db.MustBegin()
	//messageReader := NewMessageAdapter(tx)

	//messages, _ := messageReader.GetMessages()

	stream := make(chan Message)
	container.eventManager.AddTarget(stream)

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer container.eventManager.DeleteTarget(stream)
		defer close(stream)

		for {
			select {
			case message, ok := <-stream:
				if !ok {
					return
				}

				jsonData, _ := json.Marshal(message)
				strData := string(jsonData)

				_, err := fmt.Fprintf(w, "data: %s\n\n", strData)
				if err != nil {
					return
				}
				err = w.Flush()
				if err != nil {
					return
				}

			}
		}
	}))
	return nil
}
