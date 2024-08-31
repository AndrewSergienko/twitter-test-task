package internal

import (
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type IoC struct {
	db           *sqlx.DB
	conn         *amqp.Connection
	messageQueue amqp.Queue
	eventsQueue  amqp.Queue
}

func NewIoC(db *sqlx.DB, conn *amqp.Connection, mq amqp.Queue, eq amqp.Queue) *IoC {
	return &IoC{db: db, conn: conn, messageQueue: mq, eventsQueue: eq}
}

func (ioc *IoC) NewMessageAdapter() (MessageAdapter, *sqlx.Tx) {
	tx := ioc.db.MustBegin()
	return NewMessageAdapter(tx, ioc.conn, ioc.messageQueue), tx
}
