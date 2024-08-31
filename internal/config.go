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

	qMessages, _ := ch.QueueDeclare("messages", true, false, false, false, nil)
	qEvents, _ := ch.QueueDeclare("", true, false, false, false, nil)
	_ = ch.ExchangeDeclare("events", "fanout", false, false, false, false, nil)
	_ = ch.QueueBind(qEvents.Name, "", "events", false, nil)
	return &qMessages, &qEvents, nil
}
