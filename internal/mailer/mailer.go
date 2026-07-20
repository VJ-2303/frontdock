package mailer

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/smtp"
)

type Mailer struct {
	addr string
	from string
}

func New(host string, port int, from string) *Mailer {
	return &Mailer{
		addr: fmt.Sprintf("%s:%d", host, port),
		from: from,
	}
}

func (m *Mailer) Send(to, subject string, htmlBody string) error {
	msg := bytes.NewBufferString("")
	fmt.Fprintf(msg, "From: %s\r\n", m.from)
	fmt.Fprintf(msg, "To: %s\r\n", to)
	fmt.Fprintf(msg, "Subject: %s\r\n", subject)
	fmt.Fprintf(msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(msg, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(msg, "\r\n")

	fmt.Fprint(msg, htmlBody)

	err := smtp.SendMail(m.addr, nil, m.from, []string{to}, msg.Bytes())
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}
