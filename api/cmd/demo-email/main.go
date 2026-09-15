// Demo email without real data: renders or sends one SED template.
//
//	Usage (from repo root):
//	  GOPROXY=https://goproxy.io,direct go run ./api/cmd/demo-email --to test@example.com
//	  GOPROXY=https://goproxy.io,direct go run ./api/cmd/demo-email --to test@example.com --title "Demo SED" --message "Mensaje demo" --cta http://localhost:5173/assignments/demo --out /tmp/demo_email.html
//
//	Envs: MAIL_PROVIDER, SENDGRID_API_KEY, MAIL_FROM, APP_BASE_URL, DEMO_TO (or TO_EMAIL).
//	Without MAIL_PROVIDER=sendgrid + SENDGRID_API_KEY it dry-runs: renders to stdout/file and validates absolute CTA, no send.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sed-evaluacion-desempeno/api/internal/service/notify"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")
	log.Printf("demo-email: env MAIL_PROVIDER=%s MAIL_FROM=%s APP_BASE_URL=%s SENDGRID_API_KEY set=%v",
		os.Getenv("MAIL_PROVIDER"), os.Getenv("MAIL_FROM"), os.Getenv("APP_BASE_URL"), strings.TrimSpace(os.Getenv("SENDGRID_API_KEY")) != "")
	to := flag.String("to", "", "demo recipient (or DEMO_TO/TO_EMAIL)")
	title := flag.String("title", "Demo SED", "email title")
	message := flag.String("message", "Mensaje demo", "email description")
	cta := flag.String("cta", "", "absolute CTA URL (default APP_BASE_URL + /assignments/demo)")
	ctaLabel := flag.String("cta-label", "Ver demo", "CTA label")
	tplName := flag.String("template", "assignment_submitted", "assignment_submitted|change_requested")
	out := flag.String("out", "", "optional HTML output path (default stdout)")
	flag.Parse()

	tpl := notify.TemplateAssignmentSubmitted
	if *tplName == "change_requested" {
		tpl = notify.TemplateChangeRequested
	} else if *tplName != "assignment_submitted" {
		log.Fatalf("demo-email: unknown template %q", *tplName)
	}

	if *to == "" {
		*to = os.Getenv("DEMO_TO")
	}
	if *to == "" {
		*to = os.Getenv("TO_EMAIL")
	}
	baseURL := strings.TrimSuffix(strings.TrimSpace(os.Getenv("APP_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	if *cta == "" {
		*cta = baseURL + "/assignments/demo"
	}
	if !notify.IsAbsoluteURL(*cta) {
		log.Fatalf("demo-email: CTA must be absolute http(s)://, got %q", *cta)
	}

	data := map[string]string{"Title": *title, "Message": *message, "CTAURL": *cta, "CTALabel": *ctaLabel}

	apiKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("MAIL_FROM")
	renderKey, renderFrom := apiKey, from
	if renderKey == "" {
		renderKey = "dry-run" // only used for local render, never sent
	}
	if renderFrom == "" {
		renderFrom = "no-reply@example.com"
	}
	s, err := notify.NewSendGridSender(renderKey, renderFrom, baseURL)
	if err != nil {
		log.Fatalf("demo-email: sender init: %v", err)
	}
	html, err := s.RenderForTest(tpl, data)
	if err != nil {
		log.Fatalf("demo-email: render %s: %v", tpl, err)
	}
	for _, want := range []string{*title, *message, *cta, *ctaLabel} {
		if !strings.Contains(html, want) {
			log.Fatalf("demo-email: visual check failed, render missing %q (template=%s)", want, tpl)
		}
	}

	if *out != "" {
		if err := os.WriteFile(*out, []byte(html), 0o644); err != nil {
			log.Fatalf("demo-email: write out: %v", err)
		}
	} else {
		fmt.Println(html)
	}

	if os.Getenv("MAIL_PROVIDER") != "sendgrid" || strings.TrimSpace(apiKey) == "" {
		log.Printf("demo-email: dry-run ok template=%s bytes=%d (no send)", tpl, len(html))
		return
	}
	if strings.TrimSpace(*to) == "" {
		log.Fatal("demo-email: missing recipient: pass --to or set DEMO_TO")
	}
	n := notify.Notification{To: *to, Subject: *title, Template: tpl, Data: data}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Send(ctx, n); err != nil {
		log.Fatalf("demo-email: send template=%s: %v", tpl, err)
	}
	log.Printf("demo-email: sent template=%s", tpl)
}
