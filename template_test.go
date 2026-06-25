package mail

import (
	"errors"
	"html/template"
	"strings"
	"testing"
	"testing/fstest"
)

type fakeMailer struct {
	from string
	sent []Message
	err  error
}

func (f *fakeMailer) Send(msg Message) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, msg)
	return nil
}

func (f *fakeMailer) GetFrom() string {
	return f.from
}

func TestTemplateRendererFSRender(t *testing.T) {
	fsys := fstest.MapFS{
		"base.html": {
			Data: []byte(`{{define "email"}}<html><body>{{template "content" .}}</body></html>{{end}}`),
		},
		"welcome.html": {
			Data: []byte(`{{define "content"}}<h1>Hello {{.Name}}</h1>{{end}}`),
		},
	}

	renderer, err := NewTemplateRendererFS(fsys, "*.html")
	if err != nil {
		t.Fatalf("NewTemplateRendererFS() error = %v", err)
	}
	got, err := renderer.Render("email", map[string]string{"Name": "Ada"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := "<html><body><h1>Hello Ada</h1></body></html>"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestTemplateRendererEscapesHTML(t *testing.T) {
	tmpl := template.Must(template.New("email").Parse(`{{define "email"}}<p>{{.Name}}</p>{{end}}`))
	renderer, err := NewTemplateRendererFromTemplate(tmpl)
	if err != nil {
		t.Fatalf("NewTemplateRendererFromTemplate() error = %v", err)
	}

	body, err := renderer.Render("email", map[string]string{
		"Name": `<script>alert("xss")</script>`,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(body, "<script>") {
		t.Fatalf("Render() produced unescaped script tag: %q", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("Render() did not HTML-escape script input: %q", body)
	}
}

func TestTemplateRendererEscapesAttributeURLs(t *testing.T) {
	tmpl := template.Must(template.New("email").Parse(`{{define "email"}}<a href="{{.URL}}">Open</a>{{end}}`))
	renderer, err := NewTemplateRendererFromTemplate(tmpl)
	if err != nil {
		t.Fatalf("NewTemplateRendererFromTemplate() error = %v", err)
	}

	body, err := renderer.Render("email", map[string]string{
		"URL": `javascript:alert(1)`,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(body, "javascript:alert") {
		t.Fatalf("Render() allowed javascript URL: %q", body)
	}
}

func TestTemplateRendererFromTemplateMessage(t *testing.T) {
	tmpl := template.Must(template.New("email").Parse(`{{define "email"}}<p>{{.}}</p>{{end}}`))
	renderer, err := NewTemplateRendererFromTemplate(tmpl)
	if err != nil {
		t.Fatalf("NewTemplateRendererFromTemplate() error = %v", err)
	}
	msg, err := renderer.Message("Subject", "email", "Body")
	if err != nil {
		t.Fatalf("Message() error = %v", err)
	}
	if msg.GetSubject() != "Subject" {
		t.Fatalf("subject = %q", msg.GetSubject())
	}
	if msg.GetBody() != "<p>Body</p>" {
		t.Fatalf("body = %q", msg.GetBody())
	}
}

func TestTemplateRendererErrors(t *testing.T) {
	if _, err := NewTemplateRenderer(); !errors.Is(err, ErrTemplate) {
		t.Fatalf("NewTemplateRenderer() error = %v, want ErrTemplate", err)
	}
	if _, err := NewTemplateRendererFS(nil, "*.html"); !errors.Is(err, ErrTemplate) {
		t.Fatalf("NewTemplateRendererFS(nil) error = %v, want ErrTemplate", err)
	}
	if _, err := NewTemplateRendererFromTemplate(nil); !errors.Is(err, ErrTemplate) {
		t.Fatalf("NewTemplateRendererFromTemplate(nil) error = %v, want ErrTemplate", err)
	}
	if _, err := (&TemplateRenderer{}).Render("email", nil); !errors.Is(err, ErrTemplate) {
		t.Fatalf("Render(unconfigured) error = %v, want ErrTemplate", err)
	}
}

func TestTemplateMailerSendTemplate(t *testing.T) {
	tmpl := template.Must(template.New("email").Parse(`{{define "email"}}<p>Hello {{.Name}}</p>{{end}}`))
	renderer, err := NewTemplateRendererFromTemplate(tmpl)
	if err != nil {
		t.Fatalf("NewTemplateRendererFromTemplate() error = %v", err)
	}
	fake := &fakeMailer{from: "sender@example.com"}
	tm, err := NewTemplateMailer(fake, renderer)
	if err != nil {
		t.Fatalf("NewTemplateMailer() error = %v", err)
	}

	attachment := NewAttachment("hello.txt", []byte("hello"), "text/plain")
	err = tm.SendTemplate(TemplateMessage{
		Template:    "email",
		Subject:     "Welcome",
		Data:        map[string]string{"Name": "Ada"},
		To:          []string{"ada@example.com"},
		FromName:    "Support",
		Attachments: []Attachment{attachment},
	})
	if err != nil {
		t.Fatalf("SendTemplate() error = %v", err)
	}
	if len(fake.sent) != 1 {
		t.Fatalf("sent count = %d, want 1", len(fake.sent))
	}
	msg := fake.sent[0]
	if msg.GetSubject() != "Welcome" {
		t.Fatalf("subject = %q", msg.GetSubject())
	}
	if msg.GetBody() != "<p>Hello Ada</p>" {
		t.Fatalf("body = %q", msg.GetBody())
	}
	if got := msg.GetTo(); len(got) != 1 || got[0] != "ada@example.com" {
		t.Fatalf("recipients = %v", got)
	}
	if msg.GetFromName() != "Support" {
		t.Fatalf("from name = %q", msg.GetFromName())
	}
	if got := msg.GetAttachments(); len(got) != 1 || got[0].GetFilename() != "hello.txt" {
		t.Fatalf("attachments = %v", got)
	}
}

func TestTemplateMailerValidationAndSendErrors(t *testing.T) {
	tmpl := template.Must(template.New("email").Parse(`{{define "email"}}<p>Hello</p>{{end}}`))
	renderer, err := NewTemplateRendererFromTemplate(tmpl)
	if err != nil {
		t.Fatalf("NewTemplateRendererFromTemplate() error = %v", err)
	}

	if _, err := NewTemplateMailer(nil, renderer); !errors.Is(err, ErrTemplate) {
		t.Fatalf("NewTemplateMailer(nil) error = %v, want ErrTemplate", err)
	}
	if _, err := NewTemplateMailer(&fakeMailer{}, nil); !errors.Is(err, ErrTemplate) {
		t.Fatalf("NewTemplateMailer(nil renderer) error = %v, want ErrTemplate", err)
	}

	tm, err := NewTemplateMailer(&fakeMailer{}, renderer)
	if err != nil {
		t.Fatalf("NewTemplateMailer() error = %v", err)
	}
	if err := tm.SendTemplate(TemplateMessage{Template: "email", Subject: "Subject"}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("SendTemplate(missing recipient) error = %v, want ErrInvalidMessage", err)
	}

	sendErr := errors.New("smtp down")
	tm, err = NewTemplateMailer(&fakeMailer{err: sendErr}, renderer)
	if err != nil {
		t.Fatalf("NewTemplateMailer() error = %v", err)
	}
	err = tm.SendTemplate(TemplateMessage{
		Template: "email",
		Subject:  "Subject",
		To:       []string{"user@example.com"},
	})
	if !errors.Is(err, sendErr) {
		t.Fatalf("SendTemplate() error = %v, want %v", err, sendErr)
	}
}

func TestTemplateMailerRejectsHeaderInjection(t *testing.T) {
	tmpl := template.Must(template.New("email").Parse(`{{define "email"}}<p>Hello</p>{{end}}`))
	renderer, err := NewTemplateRendererFromTemplate(tmpl)
	if err != nil {
		t.Fatalf("NewTemplateRendererFromTemplate() error = %v", err)
	}
	tm, err := NewTemplateMailer(&fakeMailer{}, renderer)
	if err != nil {
		t.Fatalf("NewTemplateMailer() error = %v", err)
	}

	err = tm.SendTemplate(TemplateMessage{
		Template: "email",
		Subject:  "Subject\r\nBcc: attacker@example.com",
		To:       []string{"user@example.com"},
	})
	if !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("SendTemplate() error = %v, want ErrInvalidMessage", err)
	}
}

func TestTemplateRendererMissingTemplate(t *testing.T) {
	tmpl := template.Must(template.New("email").Parse(`{{define "email"}}<p>Hello</p>{{end}}`))
	renderer, err := NewTemplateRendererFromTemplate(tmpl)
	if err != nil {
		t.Fatalf("NewTemplateRendererFromTemplate() error = %v", err)
	}
	_, err = renderer.Render("missing", nil)
	if !errors.Is(err, ErrTemplate) {
		t.Fatalf("Render(missing) error = %v, want ErrTemplate", err)
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("Render(missing) error = %v, want template name", err)
	}
}
