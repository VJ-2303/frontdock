package queue

import (
	"context"
	"errors"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(ctx context.Context, body []byte) error

type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func Permanent(err error) error {
	return &PermanentError{Err: err}
}

func Consume(ctx context.Context, url, queueName string, prefetch int, fn HandlerFunc) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := Declare(ch); err != nil {
		return err
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false, false, false, nil,
	)
	if err != nil {
		return err
	}
	slog.Info("Consuming", "queue", queueName, "prefetch", prefetch)

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return amqp.ErrClosed
			}
			err := fn(ctx, msg.Body)

			switch {
			case err == nil:
				_ = msg.Ack(false)
			case isPermanent(err):
				slog.Error("permanent Failure, dead-lettering", "queue", queueName, "err", err)
				_ = msg.Nack(false, false)
			default:
				slog.Warn("transient failure, will retry", "queue", queueName)
				_ = msg.Nack(false, true)
			}
		}
	}
}

func isPermanent(err error) bool {
	var p *PermanentError
	return errors.As(err, &p)
}
