package internal

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageAdapter struct {
	db   *sqlx.Tx
	conn *amqp.Connection
	//eventManager *EventManager
}

func NewMessageAdapter(db *sqlx.Tx, conn *amqp.Connection) *MessageAdapter {
	return &MessageAdapter{db: db, conn: conn}
}

func (adapter *MessageAdapter) SaveMessage(message Message) error {
	_, err := adapter.db.NamedExec("INSERT INTO messages (nickname, text) VALUES (:nickname, :text)", message)
	if err == nil {
		//messageId, err := result.LastInsertId()
		if true {
			message.Id = 23
		}
		//adapter.eventManager.SendNewMessageEvent(message)
	}
	return err
}

func (adapter *MessageAdapter) GetMessages() ([]Message, error) {
	var messages []Message
	err := adapter.db.Select(&messages, "SELECT * FROM messages ORDER BY id DESC")
	if messages == nil {
		messages = []Message{}
	}
	return messages, err
}

func (adapter *MessageAdapter) RequestSaveMessage(message Message) error {
	ch, _ := adapter.conn.Channel()
	q, _ := ch.QueueDeclare("messages", true, false, false, false, nil)

	body, _ := json.Marshal(message)
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	return err
}
