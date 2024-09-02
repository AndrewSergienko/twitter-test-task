package internal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"os"
)

type DBSettings struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type MQSettings struct {
	Host     string
	Port     string
	User     string
	Password string
}

func FetchEnv(name string, strict bool) string { // nolint: all
	value := os.Getenv(name)
	if value == "" {
		if strict {
			slog.Error(fmt.Sprintf("Environment variable %s is not set", name))
			panic(fmt.Sprintf("Environment variable %s is not set", name))
		}
		slog.Warn(fmt.Sprintf("Environment variable %s is not set", name))
	}
	slog.Debug(fmt.Sprintf("Environment variable - %s: %s", name, value))
	return value
}

func NewDBSettings() DBSettings {
	return DBSettings{
		Host:     FetchEnv("DB_HOST", true),
		Port:     FetchEnv("DB_PORT", true),
		User:     FetchEnv("COCKROACH_USER", true),
		Password: FetchEnv("COCKROACH_PASSWORD", true),
		Database: FetchEnv("COCKROACH_DATABASE", true),
	}
}

func NewMQSettings() MQSettings {
	return MQSettings{
		Host:     FetchEnv("MQ_HOST", true),
		Port:     FetchEnv("MQ_PORT", true),
		User:     FetchEnv("RABBITMQ_DEFAULT_USER", true),
		Password: FetchEnv("RABBITMQ_DEFAULT_PASS", true),
	}
}

func NewDB(settings DBSettings) (*sqlx.DB, error) {
	rootCert := "app/certs/ca.crt"
	clientCert := fmt.Sprintf("app/certs/client.%s.crt", settings.User)
	clientKey := fmt.Sprintf("app/certs/client.%s.crt", settings.User)

	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=verify-full&sslrootcert=%s&sslcert=%s&sslkey=%s",
		settings.User,
		settings.Password,
		settings.Host,
		settings.Port,
		settings.Database,
		rootCert,
		clientCert,
		clientKey,
	)
	return sqlx.Connect("postgres", connStr)
}

func NewMQConn(settings MQSettings) (*amqp.Connection, error) {
	connStr := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		settings.User,
		settings.Password,
		settings.Host,
		settings.Port,
	)
	return amqp.Dial(connStr)
}

func SetupQueues(conn *amqp.Connection) (*amqp.Queue, *amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			return
		}
	}(ch)

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
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			return
		}
	}(ch)

	_, _ = ch.QueueDelete(qMessages.Name, true, true, false)
	_, _ = ch.QueueDelete(qEvents.Name, true, true, false)
}
