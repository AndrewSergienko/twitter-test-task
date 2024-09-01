package main

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"twitter-test-task/internal"
)

func main() {
	db, err := internal.NewDB()
	if err != nil {
		slog.Error(fmt.Sprintf("Cannot connect to DB: %v", err))
		panic("Cannot connect to DB")
	}

	conn, err := internal.NewMQConn()
	if err != nil {
		slog.Error(fmt.Sprintf("Cannot connect to MQ: %v", err))
		panic("Cannot connect to MQ")
	}

	mq, eq, err := internal.SetupQueues(conn)
	if err != nil {
		slog.Error(fmt.Sprintf("Cannot setup queues: %v", err))
		panic("Cannot setup queues")
	}
	defer internal.FinalizeQueues(conn, mq, eq)

	em := internal.NewEventManager()
	ioc := internal.NewIoC(db, conn, *mq, *eq)

	worker := internal.NewWorker(conn, *mq, *eq, *ioc, em)

	app := internal.NewWebApp(*ioc, em)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)

	go func() {
		err := worker.RunWorker(ctx)
		if err != nil {
			slog.Error(fmt.Sprintf("Worker error: %v", err))
			errChan <- err
		}
	}()
	go func() {
		err := worker.RunObserver(ctx)
		if err != nil {
			slog.Error(fmt.Sprintf("Observer error: %v", err))
			errChan <- err
		}
	}()
	go func() {
		if err := app.Listen(":3000"); err != nil {
			slog.Error(fmt.Sprintf("Server error: %v", err))
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		slog.Info(fmt.Sprint("Shutting down due to error: ", err))
		cancel()
		if app.Shutdown() == nil {
			slog.Info("Server shutdown")
			return
		}
	case <-ctx.Done():
		slog.Info("Context canceled, shutting down server")
	}
}
