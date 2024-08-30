package internal

import "github.com/jmoiron/sqlx"

type MessageAdapter struct {
	db           *sqlx.Tx
	eventManager *EventManager
}

func NewMessageAdapter(db *sqlx.Tx, em *EventManager) *MessageAdapter {
	return &MessageAdapter{db: db, eventManager: em}
}

func (adapter *MessageAdapter) SaveMessage(message Message) error {
	_, err := adapter.db.NamedExec("INSERT INTO messages (nickname, text) VALUES (:nickname, :text)", message)
	if err == nil {
		//messageId, err := result.LastInsertId()
		if true {
			message.Id = 23
		}
		adapter.eventManager.SendEvent(message)
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
