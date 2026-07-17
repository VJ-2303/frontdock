package queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeMain = "frontdock"
	ExchangeDLX  = "frontdock.dlx"

	RoutingEmailSend       = "email.send"
	RoutingDeployRequested = "deploy.requested"

	QueueEmail     = "email.jobs"
	QueueDeploy    = "deploy.jobs"
	QueueEmailDLQ  = "email.jobs.dlx"
	QueueDeployDLQ = "deploy.jobs.dlx"

	MaxAttempts = 3
)

func Declare(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(ExchangeMain, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(ExchangeDLX, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	for q, rk := range map[string]string{
		QueueEmailDLQ:  RoutingEmailSend,
		QueueDeployDLQ: RoutingDeployRequested,
	} {
		if _, err := ch.QueueDeclare(q, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(q, rk, ExchangeDLX, false, nil); err != nil {
			return err
		}
	}
	args := amqp.Table{"x-dead-letter-exchange": ExchangeDLX}

	for q, rk := range map[string]string{
		QueueEmail:  RoutingEmailSend,
		QueueDeploy: RoutingDeployRequested,
	} {
		if _, err := ch.QueueDeclare(q, true, false, false, false, args); err != nil {
			return err
		}
		if err := ch.QueueBind(q, rk, ExchangeMain, false, nil); err != nil {
			return err
		}
	}
	return nil
}
