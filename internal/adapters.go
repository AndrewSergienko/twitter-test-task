package internal

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageAdapter struct {
	db           *sqlx.Tx
	conn         *amqp.Connection
	messageQueue amqp.Queue
}

func NewMessageAdapter(db *sqlx.Tx, conn *amqp.Connection, mq amqp.Queue) MessageAdapter {
	return MessageAdapter{db: db, conn: conn, messageQueue: mq}
}

func (adapter *MessageAdapter) SaveMessage(message Message) (int, error) {
	rows, err := adapter.db.NamedQuery(
		"INSERT INTO messages (nickname, text) VALUES (:nickname, :text) RETURNING id",
		message,
	)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			message.Id = id
			return id, err
		}
	}
	return 0, err
}

func (adapter *MessageAdapter) GetMessages(from int, limit int) ([]Message, error) {
	var messages []Message
	err := adapter.db.Select(&messages, "SELECT * FROM messages WHERE id >= $1 ORDER BY id DESC LIMIT $2", from, limit)
	if messages == nil {
		messages = []Message{}
	}
	return messages, err
}

func (adapter *MessageAdapter) RequestSaveMessage(message Message) error {
	ch, _ := adapter.conn.Channel()

	body, _ := json.Marshal(message)
	publishing := amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	}
	err := ch.Publish("", adapter.messageQueue.Name, false, false, publishing)
	return err
}
