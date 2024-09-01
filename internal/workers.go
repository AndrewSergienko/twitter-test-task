package internal

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
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

func (worker *Worker) RunWorker() error {
	ch, err := worker.conn.Channel()
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(worker.messagesQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	for d := range msgs {
		var data *Message
		if json.Unmarshal(d.Body, &data) != nil || data == nil {
			if d.Nack(false, false) != nil {
				continue
			}
		}

		messageAdapter, tx := worker.ioc.NewMessageAdapter()

		id, err := messageAdapter.SaveMessage(*data)
		if err != nil {
			slog.Warn("Cannot save message")
			return err
		}
		if err = tx.Commit(); err != nil {
			slog.Warn("Cannot commit transaction")
			return err
		}
		slog.Info("Message saved")

		data.Id = id

		body, err := json.Marshal(data)
		if err != nil {
			if d.Nack(false, true) != nil {
				continue
			}
		}

		publishing := amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}
		if ch.Publish("events", "", false, false, publishing) != nil {
			if d.Nack(false, true) != nil {
				continue
			}
		} else {
			slog.Info("Message published")
		}
	}

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			return
		}
	}(ch)
	return nil
}

func (worker *Worker) RunObserver() error {
	ch, err := worker.conn.Channel()
	if err != nil {
		return err
	}
	msgs, _ := ch.Consume(worker.eventsQueue.Name, "", true, false, false, false, nil)
	for msg := range msgs {
		slog.Info("New message received")
		var message Message
		err := json.Unmarshal(msg.Body, &message)
		if err != nil {
			continue
		}

		worker.eventManager.SendNewMessageEvent(message)
	}
	return nil
}
