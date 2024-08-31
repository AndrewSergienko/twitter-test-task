package main

import (
	"log"
	"log/slog"
	"twitter-test-task/internal"
)

func main() {
	db, err := internal.NewDB()
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := internal.NewMQConn()
	if err != nil {
		log.Fatalln(err)
	}

	mq, eq, err := internal.SetupQueues(conn)
	if err != nil {
		log.Fatalln(err)
	}

	em := internal.NewEventManager()
	ioc := internal.NewIoC(db, conn, *mq, *eq)

	worker := internal.NewWorker(conn, *mq, *eq, *ioc, em)

	app := internal.NewWebApp(*ioc, em)

	go worker.RunWorker()
	go worker.RunObserver()

	slog.Error(app.Listen(":3000").Error())
}
