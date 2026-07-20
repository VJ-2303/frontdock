package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/VJ-2303/frontdock/internal/config"
	"github.com/VJ-2303/frontdock/internal/database"
	"github.com/VJ-2303/frontdock/internal/mailer"
	"github.com/VJ-2303/frontdock/internal/queue"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	godotenv.Load()
	cfg, err := config.Load(config.ServiceWorker)
	if err != nil {
		return err
	}

	pool, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	mail := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.MailFrom)

	pub, err := queue.NewPublisher(cfg.RabbitURL)
	if err != nil {
		return err
	}
	defer pub.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Info("worker shutting down")
		cancel()
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		runForever(ctx, "email", func() error {
			return queue.Consume(ctx, cfg.RabbitURL, queue.QueueEmail, 10, emailHandler(mail, cfg))
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runForever(ctx, "deploy", func() error {
			return queue.Consume(ctx, cfg.RabbitURL, queue.QueueDeploy, 1, func(ctx context.Context, body []byte) error {
				slog.Info("deploy job received (not implemented yet)")
				return nil
			})
		})
	}()

	wg.Wait()
	return nil
}

func runForever(ctx context.Context, name string, f func() error) {
	for {
		if err := f(); err != nil {
			slog.Error("consumer died, reconnecting in 5s", "consumer", name, "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
