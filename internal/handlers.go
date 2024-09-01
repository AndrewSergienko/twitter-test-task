package internal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"strconv"
	"time"
)

type MessageHandlers struct {
	eventManager *EventManager
	ioc          IoC
}

func NewMessageHandlers(em *EventManager, ioc IoC) *MessageHandlers {
	return &MessageHandlers{
		eventManager: em,
		ioc:          ioc,
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

	messageAdapter, _ := container.ioc.NewMessageAdapter()

	message := Message{
		Nickname: requestData.Nickname,
		Text:     requestData.Text,
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if messageAdapter.RequestSaveMessage(ctx, message) != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot create request"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok"})
}

func (container MessageHandlers) GetMessages(c *fiber.Ctx) error {
	live, _ := strconv.ParseBool(c.Query("live", "false"))
	from, _ := strconv.Atoi(c.Query("from", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	if live && from != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot use 'live' and 'from' parameters together"})
	}

	messageReader, _ := container.ioc.NewMessageAdapter()

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	messages, err := messageReader.GetMessages(ctx, from, limit)
	if err != nil {
		slog.Warn(fmt.Sprintf("Cannot get messages: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot get messages"})
	}

	if !live {
		return c.JSON(messages)
	}

	container.runMessageStreamer(c, messages)
	return nil
}

func (container MessageHandlers) runMessageStreamer(c *fiber.Ctx, messages []Message) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	stream := make(chan Message)
	container.eventManager.AddTarget(stream)

	closeNotify := c.Context().Done()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		keepAliveTicker := time.NewTicker(10 * time.Second)
		keepAliveMsg := ":keepalive\n"

		defer container.eventManager.DeleteTarget(stream)
		defer close(stream)

		jsonData, _ := json.Marshal(messages)
		if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
			return
		}
		if w.Flush() != nil {
			return
		}

		for {
			select {
			case message, ok := <-stream:
				if !ok {
					return
				}

				jsonData, _ := json.Marshal([]Message{message})
				strData := string(jsonData)

				if _, err := fmt.Fprintf(w, "data: %s\n\n", strData); err != nil {
					return
				}
				if w.Flush() != nil {
					return
				}
			case <-keepAliveTicker.C:
				if _, err := fmt.Fprintf(w, keepAliveMsg); err != nil {
					slog.Info("Client disconnected. Stopped request")
					return
				}
				if w.Flush() != nil {
					slog.Info("Client disconnected. Stopped request")
					return
				}
			case <-closeNotify:
				return
			}
		}
	})
}
