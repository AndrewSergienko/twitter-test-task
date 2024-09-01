package internal

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"time"
)

type Worker struct {
	conn          *amqp.Connection
	messagesQueue amqp.Queue
	eventsQueue   amqp.Queue
	eventManager  *EventManager
	ioc           IoC
}

func NewWorker(conn *amqp.Connection, mq amqp.Queue, eq amqp.Queue, ioc IoC, em *EventManager) Worker {
	return Worker{conn: conn, messagesQueue: mq, eventsQueue: eq, ioc: ioc, eventManager: em}
}

func (worker *Worker) RunWorker(ctx context.Context) error {
	ch, err := worker.conn.Channel()
	if err != nil {
		return err
	}
	defer func(ch *amqp.Channel) {
		if err := ch.Close(); err != nil {
			return
		}
	}(ch)

	msgs, err := ch.Consume(worker.messagesQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	processMessage := func(d amqp.Delivery) error {
		var data *Message
		if json.Unmarshal(d.Body, &data) != nil || data == nil {
			_ = d.Nack(false, false)
			return nil
		}

		messageAdapter, tx := worker.ioc.NewMessageAdapter()
		ctxDB, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := messageAdapter.SaveMessage(ctxDB, *data)
		if err != nil {
			slog.Warn("Cannot save message")
			return err
		}
		if err = tx.Commit(); err != nil {
			slog.Warn("Cannot commit transaction")
			return err
		}
		slog.Info("Message saved")

		data.ID = id
		body, err := json.Marshal(data)
		if err != nil {
			_ = d.Nack(false, false)
			return nil
		}

		ctxMQ, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		publishing := amqp.Publishing{ContentType: "application/json", Body: body}
		if ch.PublishWithContext(ctxMQ, "events", "", false, false, publishing) != nil {
			slog.Warn(fmt.Sprintf("Cannot publish message: %v", err))
			_ = d.Nack(false, false)
			return nil
		}
		slog.Info("Message published")
		return nil
	}

	slog.Info("Start worker")
	for {
		select {
		case d := <-msgs:
			if err := processMessage(d); err != nil {
				return err
			}
		case <-ctx.Done():
			slog.Info("Shutdown worker")
			return nil
		}
	}
}

func (worker *Worker) RunObserver(ctx context.Context) error {
	ch, err := worker.conn.Channel()
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(worker.eventsQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("Start observer")
	for {
		select {
		case msg := <-msgs:
			if msg.Body == nil {
				continue
			}

			slog.Info("New message received")
			var message Message
			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				slog.Warn("Cannot unmarshal event")
				if err := msg.Nack(false, false); err != nil {
					continue
				}
			}

			worker.eventManager.SendNewMessageEvent(message)
		case <-ctx.Done():
			slog.Info("Shutdown observer")
			if err := ch.Close(); err != nil {
				return err
			}
			return nil
		}
	}
}
