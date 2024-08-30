package internal

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Worker struct {
	conn *amqp.Connection
	//messageSaver MessageSaver
	eventManager *EventManager
	db           *sqlx.DB
}

func NewWorker(conn *amqp.Connection, db *sqlx.DB, em *EventManager) Worker {
	return Worker{conn: conn, db: db, eventManager: em}
}

func (worker *Worker) RunWorker() error {
	ch, err := worker.conn.Channel()
	if err != nil {
		return err
	}
	q, err := ch.QueueDeclare("messages", true, false, false, false, nil)
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	for d := range msgs {
		var data *Message
		err := json.Unmarshal(d.Body, &data)
		if err != nil {
			continue
		}
		if data != nil {
			tx := worker.db.MustBegin()
			messageSaver := NewMessageAdapter(tx, worker.conn)

			err := messageSaver.SaveMessage(*data)
			_ = tx.Commit()
			if err != nil {
				return err
			}
			body, _ := json.Marshal(data)
			_ = ch.Publish(
				"events", // exchange
				"",       // routing key
				false,    // mandatory
				false,    // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        body,
				})
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
	ch, _ := worker.conn.Channel()
	_ = ch.ExchangeDeclare("events", "fanout", false, false, false, false, nil)

	queue, _ := ch.QueueDeclare(
		"",    // Назва черги (авто-генерація)
		false, // Стійка
		false, // Авто-видалення
		true,  // Вимагати виключення
		false, // Не чекати
		nil,   // Додаткові параметри
	)

	_ = ch.QueueBind(
		queue.Name,
		"",
		"events",
		false,
		nil,
	)

	msgs, _ := ch.Consume(
		queue.Name,
		"",   // Споживач не ідентифікований
		true, // Авто-підтвердження
		false,
		false,
		false,
		nil,
	)
	for msg := range msgs {
		var message Message
		_ = json.Unmarshal(msg.Body, &message)

		worker.eventManager.SendNewMessageEvent(message)
	}
	return nil
}
