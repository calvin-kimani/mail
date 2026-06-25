package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
)

// ErrTemplate is returned when template parsing, rendering, or sending fails.
var ErrTemplate = errors.New("mail: template error")

// TemplateRenderer renders HTML email bodies from html/template templates.
type TemplateRenderer struct {
	templates *template.Template
}

// NewTemplateRenderer parses templates from filesystem paths.
func NewTemplateRenderer(filenames ...string) (*TemplateRenderer, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("%w: at least one template file is required", ErrTemplate)
	}
	tmpl, err := template.ParseFiles(filenames...)
	if err != nil {
		return nil, fmt.Errorf("%w: parse files: %w", ErrTemplate, err)
	}
	return &TemplateRenderer{templates: tmpl}, nil
}

// NewTemplateRendererFS parses templates from fsys using html/template.ParseFS.
func NewTemplateRendererFS(fsys fs.FS, patterns ...string) (*TemplateRenderer, error) {
	if fsys == nil {
		return nil, fmt.Errorf("%w: fs is required", ErrTemplate)
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("%w: at least one template pattern is required", ErrTemplate)
	}
	tmpl, err := template.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, fmt.Errorf("%w: parse fs: %w", ErrTemplate, err)
	}
	return &TemplateRenderer{templates: tmpl}, nil
}

// NewTemplateRendererFromTemplate wraps an already parsed template set.
func NewTemplateRendererFromTemplate(tmpl *template.Template) (*TemplateRenderer, error) {
	if tmpl == nil {
		return nil, fmt.Errorf("%w: template is required", ErrTemplate)
	}
	return &TemplateRenderer{templates: tmpl}, nil
}

// Render executes the named template and returns the rendered HTML string.
func (r *TemplateRenderer) Render(name string, data any) (string, error) {
	if r == nil || r.templates == nil {
		return "", fmt.Errorf("%w: renderer is not configured", ErrTemplate)
	}
	if name == "" {
		return "", fmt.Errorf("%w: template name is required", ErrTemplate)
	}
	var body bytes.Buffer
	if err := r.templates.ExecuteTemplate(&body, name, data); err != nil {
		return "", fmt.Errorf("%w: render %q: %w", ErrTemplate, name, err)
	}
	return body.String(), nil
}

// Message renders a template and returns a Message with the rendered HTML body.
func (r *TemplateRenderer) Message(subject, templateName string, data any) (Message, error) {
	body, err := r.Render(templateName, data)
	if err != nil {
		return nil, err
	}
	return NewMessage(subject, body), nil
}

// TemplateMailer sends rendered template emails through an underlying Mailer.
type TemplateMailer struct {
	Mailer   Mailer
	Renderer *TemplateRenderer
}

// TemplateMessage contains the data required to render and send a template email.
type TemplateMessage struct {
	Template    string
	Subject     string
	Data        any
	To          []string
	FromName    string
	Attachments []Attachment
}

// NewTemplateMailer creates a template-aware wrapper around a Mailer.
func NewTemplateMailer(mailer Mailer, renderer *TemplateRenderer) (*TemplateMailer, error) {
	if mailer == nil {
		return nil, fmt.Errorf("%w: mailer is required", ErrTemplate)
	}
	if renderer == nil {
		return nil, fmt.Errorf("%w: renderer is required", ErrTemplate)
	}
	return &TemplateMailer{Mailer: mailer, Renderer: renderer}, nil
}

// SendTemplate renders req.Template with req.Data and sends it.
func (m *TemplateMailer) SendTemplate(req TemplateMessage) error {
	if m == nil || m.Mailer == nil {
		return fmt.Errorf("%w: mailer is not configured", ErrTemplate)
	}
	if m.Renderer == nil {
		return fmt.Errorf("%w: renderer is not configured", ErrTemplate)
	}
	msg, err := m.Renderer.Message(req.Subject, req.Template, req.Data)
	if err != nil {
		return err
	}
	for _, recipient := range req.To {
		msg.AddRecipient(recipient)
	}
	if req.FromName != "" {
		msg.SetFromName(req.FromName)
	}
	for _, attachment := range req.Attachments {
		msg.AddAttachment(attachment)
	}
	if err := ValidateMessage(msg); err != nil {
		return err
	}
	return m.Mailer.Send(msg)
}
