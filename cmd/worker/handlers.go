package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/VJ-2303/frontdock/internal/config"
	"github.com/VJ-2303/frontdock/internal/mailer"
	"github.com/VJ-2303/frontdock/internal/queue"
)

func emailHandler(m *mailer.Mailer, cfg *config.Config) queue.HandlerFunc {
	return func(ctx context.Context, body []byte) error {
		var msg queue.EmailMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			return queue.Permanent(fmt.Errorf("malformed email message: %w", err))
		}
		subject, html, err := mailer.Render(msg.Template, msg.Data, cfg.SiteDomain, cfg.PublicAPIURL)
		if err != nil {
			return queue.Permanent(err)
		}

		return m.Send(msg.To, subject, html)
	}
}
