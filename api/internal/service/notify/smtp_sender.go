package notify

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
	"net/smtp"
	"html/template"
	"time"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Config for the SMTP transport. Loaded from env in main.go.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string        // always the generic no-reply address
	Timeout  time.Duration // per-send deadline
}

type SMTPSender struct {
	cfg      Config
	rendered map[Template]*template.Template
}

func NewSMTPSender(cfg Config) (*SMTPSender, error) {
	s := &SMTPSender{cfg: cfg, rendered: make(map[Template]*template.Template)}
	for _, name := range []Template{TemplateCommentCreated} {
		body, err := templateFS.ReadFile("templates/" + string(name) + ".tmpl")
		if err != nil {
			return nil, fmt.Errorf("notify: load %s: %w", name, err)
		}
		t, err := template.New(string(name)).Parse(string(body))
		if err != nil {
			return nil, fmt.Errorf("notify: parse %s: %w", name, err)
		}
		s.rendered[name] = t
	}
	return s, nil
}

func (s *SMTPSender) Send(ctx context.Context, n Notification) error {
	t, ok := s.rendered[n.Template]
	if !ok {
		return fmt.Errorf("notify: unknown template %q", n.Template)
	}
	var body bytes.Buffer
	if err := t.Execute(&body, n.Data); err != nil {
		return fmt.Errorf("notify: render %s: %w", n.Template, err)
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.cfg.From, n.To, n.Subject, body.String(),
	))

	// ponytail: net/smtp ignores ctx; goroutine + select bounds it.
	// Upgrade to gomail when we need real TLS/auth nuance.
	errCh := make(chan error, 1)
	go func() { errCh <- smtp.SendMail(addr, auth, s.cfg.From, []string{n.To}, msg) }()
	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("notify.send.failed template=%s to=%s err=%v", n.Template, n.To, err)
		}
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
