package notify

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func newTestSender(t *testing.T) *SendGridSender {
	t.Helper()
	s, err := NewSendGridSender("test-key", "no-reply@example.com", "https://app.example.com")
	if err != nil {
		t.Fatalf("NewSendGridSender: %v", err)
	}
	return s
}

func TestRenderAssignmentSubmitted(t *testing.T) {
	s := newTestSender(t)
	data := map[string]string{
		"Title":    "Tu asignación está lista",
		"Message":  "Revisa tus metas del ciclo actual.",
		"CTAURL":   "https://app.example.com/objetivos/asignacion",
		"CTALabel": "Ver mi asignación",
	}
	html, err := s.RenderForTest(TemplateAssignmentSubmitted, data)
	if err != nil {
		t.Fatalf("RenderForTest: %v", err)
	}
	for _, want := range []string{data["Title"], data["Message"], data["CTAURL"], data["CTALabel"], "No responder"} {
		if !strings.Contains(html, want) {
			t.Errorf("render missing %q", want)
		}
	}
	// Idempotencia: mismo input → mismo output.
	again, err := s.RenderForTest(TemplateAssignmentSubmitted, data)
	if err != nil {
		t.Fatalf("RenderForTest again: %v", err)
	}
	if html != again {
		t.Error("render not deterministic")
	}
}

func TestRenderChangeRequested(t *testing.T) {
	s := newTestSender(t)
	data := map[string]string{
		"Title":    "Solicitaron cambios en tus metas",
		"Message":  "Tu jefe revisó tu asignación y propone cambios.",
		"CTAURL":   "https://app.example.com/objetivos/asignacion",
		"CTALabel": "Ver mi asignación",
	}
	html, err := s.RenderForTest(TemplateChangeRequested, data)
	if err != nil {
		t.Fatalf("RenderForTest: %v", err)
	}
	for _, want := range []string{data["Title"], data["Message"], data["CTAURL"], data["CTALabel"]} {
		if !strings.Contains(html, want) {
			t.Errorf("render missing %q", want)
		}
	}
}

type captureSender struct {
	mu  sync.Mutex
	got []Notification
}

func (c *captureSender) Send(_ context.Context, n Notification) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.got = append(c.got, n)
	return nil
}

// Ambos eventos del change llegan al Sender con CTA absoluta.
func TestBothEventsReachSender(t *testing.T) {
	cap := &captureSender{}
	ctx := context.Background()
	events := []Notification{
		{
			To:       "colab@example.com",
			Subject:  "Nueva asignación",
			Template: TemplateAssignmentSubmitted,
			Data: map[string]string{
				"Title": "Tu asignación está lista", "Message": "Revisa tus metas.",
				"CTAURL": "https://app.example.com/objetivos/asignacion", "CTALabel": "Ver mi asignación",
			},
		},
		{
			To:       "colab@example.com",
			Subject:  "Tu jefe solicitó cambios en tus metas",
			Template: TemplateChangeRequested,
			Data: map[string]string{
				"Title": "Solicitaron cambios en tus metas", "Message": "Revisa el detalle.",
				"CTAURL": "https://app.example.com/objetivos/asignacion", "CTALabel": "Ver mi asignación",
			},
		},
	}
	for _, n := range events {
		if err := cap.Send(ctx, n); err != nil {
			t.Fatalf("Send: %v", err)
		}
	}
	if len(cap.got) != 2 {
		t.Fatalf("got %d notifications, want 2", len(cap.got))
	}
	for _, n := range cap.got {
		if !IsAbsoluteURL(n.Data["CTAURL"]) {
			t.Errorf("template %s: CTA no absoluta: %q", n.Template, n.Data["CTAURL"])
		}
	}
	if cap.got[0].Template != TemplateAssignmentSubmitted || cap.got[1].Template != TemplateChangeRequested {
		t.Errorf("unexpected templates: %q, %q", cap.got[0].Template, cap.got[1].Template)
	}
}
