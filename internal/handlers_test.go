package internal

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type HandlersTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	conn *amqp.Connection
	mq   amqp.Queue
	eq   amqp.Queue
	app  *fiber.App
}

func (suite *HandlersTestSuite) SetupSuite() {
	db, err := NewDB()
	suite.NoError(err)
	suite.db = db

	conn, err := NewMQConn()
	suite.NoError(err)
	suite.conn = conn

	mq, eq, err := SetupQueues(conn)
	suite.NoError(err)
	suite.mq = *mq
	suite.eq = *eq

	ioc := NewIoC(db, conn, *mq, *eq)
	em := NewEventManager()
	worker := NewWorker(conn, *mq, *eq, *ioc, em)

	go suite.NoError(worker.RunWorker(context.Background()))

	suite.app = NewWebApp(*ioc, em)
}

func (suite *HandlersTestSuite) SetupTest() {
	_, err := suite.db.Exec("TRUNCATE messages")
	suite.NoError(err)
}

func (suite *HandlersTestSuite) TestCreateMessage() {
	req := httptest.NewRequest("POST", "/api/messages/", strings.NewReader(`{"nickname":"test","text":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(201, resp.StatusCode)

	ch, err := suite.conn.Channel()
	suite.NoError(err)
	defer suite.NoError(ch.Close())

	msgs, err := ch.Consume(suite.eq.Name, "", true, false, false, false, nil)
	suite.NoError(err)

	timeout := 5 * time.Second

	for {
		select {
		case <-msgs:
			messages, err := suite.db.Query("SELECT nickname, text FROM messages")

			count := 0
			for messages.Next() {
				count++

				var nickname string
				var text string

				err = messages.Scan(&nickname, &text)
				suite.NoError(err)

				suite.Equal("test", nickname)
				suite.Equal("test", text)
			}
			suite.Equal(1, count)
			suite.NoError(err)
			return
		case <-time.After(timeout):
			return
		}
	}
}

func (suite *HandlersTestSuite) TestGetMessages() {
	tx := suite.db.MustBegin()
	tx.MustExec("INSERT INTO messages (nickname, text) VALUES ('test', 'test')")
	err := tx.Commit()
	suite.NoError(err)

	req := httptest.NewRequest("GET", "/api/messages/", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)

	type message struct {
		Nickname string `json:"nickname"`
		Text     string `json:"text"`
	}
	var messages []message

	suite.NoError(json.Unmarshal(body, &messages))
	suite.Len(messages, 1)
	suite.Equal("test", messages[0].Nickname)
	suite.Equal("test", messages[0].Text)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
