package internal

import (
	"context"
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

func (adapter *MessageAdapter) SaveMessage(ctx context.Context, message Message) (int, error) {
	query := "INSERT INTO messages (nickname, text) VALUES (:nickname, :text) RETURNING id"
	stmt, err := adapter.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return 0, err
	}
	rows, err := stmt.QueryContext(ctx, message)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			message.ID = id
			return id, err
		}
	}
	return 0, err
}

func (adapter *MessageAdapter) GetMessages(ctx context.Context, from int, limit int) ([]Message, error) {
	var messages []Message
	stmt := "SELECT * FROM messages WHERE id >= :from ORDER BY id DESC LIMIT :limit"
	err := adapter.db.SelectContext(ctx, &messages, stmt, from, limit)
	if messages == nil {
		messages = []Message{}
	}
	return messages, err
}

func (adapter *MessageAdapter) RequestSaveMessage(ctx context.Context, message Message) error {
	ch, _ := adapter.conn.Channel()

	body, _ := json.Marshal(message)
	publishing := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}
	err := ch.PublishWithContext(ctx, "", adapter.messageQueue.Name, false, false, publishing)
	return err
}
