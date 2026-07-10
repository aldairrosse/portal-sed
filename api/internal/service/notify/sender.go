// Package notify ships outbound notifications (v1: email only).
package notify

import "context"

// Template is a known template name; keeps callers honest.
type Template string

const (
	TemplateCommentCreated Template = "comment_created"
)

// Notification is the rendered payload handed to a Sender. Keep it flat and
// small — everything the template needs, nothing more.
type Notification struct {
	To       string
	Subject  string
	Template Template
	Data     map[string]string
}

// Sender renders and ships a notification. Implementations must honor ctx.
type Sender interface {
	Send(ctx context.Context, n Notification) error
}

// NoopSender discards every notification. Use in dev and tests.
type NoopSender struct{}

func (NoopSender) Send(_ context.Context, _ Notification) error { return nil }
