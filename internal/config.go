package internal

import (
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewDB() (*sqlx.DB, error) {
	connStr := "postgresql://user1:password@localhost:26257/database?sslmode=disable"
	return sqlx.Connect("postgres", connStr)
}

func NewMQConn() (*amqp.Connection, error) {
	return amqp.Dial("amqp://user:password@localhost:5672/")
}

func SetupQueues(conn *amqp.Connection) (*amqp.Queue, *amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	defer ch.Close()

	qMessages, err := ch.QueueDeclare("messages", true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}
	qEvents, err := ch.QueueDeclare("", true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}
	if err = ch.ExchangeDeclare("events", "fanout", false, false, false, false, nil); err != nil {
		return nil, nil, err
	}
	if err = ch.QueueBind(qEvents.Name, "", "events", false, nil); err != nil {
		return nil, nil, err
	}
	return &qMessages, &qEvents, nil
}

func FinalizeQueues(conn *amqp.Connection, qMessages *amqp.Queue, qEvents *amqp.Queue) {
	ch, _ := conn.Channel()
	defer ch.Close()

	_, _ = ch.QueueDelete(qMessages.Name, true, true, false)
	_, _ = ch.QueueDelete(qEvents.Name, true, true, false)
}
