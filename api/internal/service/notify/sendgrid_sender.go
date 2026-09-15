package notify

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridSender ships notifications via SendGrid v3 API behind Sender.
type SendGridSender struct {
	client  *sendgrid.Client
	from    string
	baseURL string
	render  map[Template]*template.Template
}

// NewSendGridSender builds a sender; empty apiKey returns an error so callers fall back to NoopSender.
func NewSendGridSender(apiKey, from, baseURL string) (*SendGridSender, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("notify: missing SENDGRID_API_KEY")
	}
	if strings.TrimSpace(from) == "" {
		return nil, fmt.Errorf("notify: missing MAIL_FROM")
	}
	s := &SendGridSender{client: sendgrid.NewSendClient(apiKey), from: from, baseURL: baseURL, render: make(map[Template]*template.Template)}
	for _, name := range []Template{TemplateCommentCreated, TemplateAssignmentSubmitted, TemplateChangeRequested} {
		body, err := templateFS.ReadFile("templates/" + string(name) + ".tmpl")
		if err != nil {
			return nil, fmt.Errorf("notify: load %s: %w", name, err)
		}
		t, err := template.New(string(name)).Parse(string(body))
		if err != nil {
			return nil, fmt.Errorf("notify: parse %s: %w", name, err)
		}
		s.render[name] = t
	}
	return s, nil
}

// IsAbsoluteURL reports whether u is an absolute http(s) URL for email CTAs.
func IsAbsoluteURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}

func (s *SendGridSender) Send(ctx context.Context, n Notification) error {
	t, ok := s.render[n.Template]
	if !ok {
		return fmt.Errorf("notify: unknown template %q", n.Template)
	}
	if cta := n.Data["CTAURL"]; cta != "" && !IsAbsoluteURL(cta) {
		return fmt.Errorf("notify: CTAURL must be absolute https?://, got %q", cta)
	}
	var body bytes.Buffer
	if err := t.Execute(&body, n.Data); err != nil {
		return fmt.Errorf("notify: render %s: %w", n.Template, err)
	}
	from := mail.NewEmail("", s.from)
	to := mail.NewEmail("", n.To)
	msg := mail.NewSingleEmail(from, n.Subject, to, body.String(), body.String())
	resp, err := s.client.SendWithContext(ctx, msg)
	if err != nil {
		log.Printf("notify.send.failed template=%s err=%v", n.Template, err)
		return err
	}
	if resp.StatusCode >= 300 {
		log.Printf("notify.send.failed template=%s status=%d body=%.200s", n.Template, resp.StatusCode, resp.Body)
		return fmt.Errorf("notify: sendgrid status %d", resp.StatusCode)
	}
	return nil
}

// RenderForTest renders a template to HTML for demo/visual tests without sending.
// ponytail: exported only for the demo + visual test; not part of Sender.
func (s *SendGridSender) RenderForTest(tpl Template, data map[string]string) (string, error) {
	t, ok := s.render[tpl]
	if !ok {
		return "", fmt.Errorf("notify: unknown template %q", tpl)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
