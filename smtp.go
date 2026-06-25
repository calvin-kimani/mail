package mail

import (
	"io"

	"gopkg.in/gomail.v2"
)

// Ensure SMTPMailer implements the Mailer interface at compile time.
var _ Mailer = &SMTPMailer{}

// SMTPMailer is a generic SMTP mail service implementation that can work with any SMTP server.
// It provides flexibility to connect to various email providers like Gmail, Outlook, Yahoo, or custom SMTP servers.
type SMTPMailer struct {
	Host  string // SMTP server hostname (e.g., "smtp.gmail.com", "smtp.office365.com")
	Port  int    // SMTP server port (typically 587 for TLS, 465 for SSL, or 25 for unencrypted)
	Email string // SMTP email/username (usually the email address)
	Pass  string // SMTP password or app-specific password
}

// NewSMTPMailer creates a new SMTP mailer from explicit connection settings:
//   - host: SMTP server hostname (e.g., "smtp.gmail.com", "smtp.office365.com")
//   - port: SMTP server port (typically 587 for STARTTLS, 465 for SSL)
//   - email: SMTP username, usually the sender's email address
//   - pass: SMTP password or app-specific password
//
// Common SMTP configurations:
//   - Gmail: smtp.gmail.com:587 (requires app password)
//   - Outlook: smtp-mail.outlook.com:587
//   - Yahoo: smtp.mail.yahoo.com:587
//
// Resolving these values (from the environment, a config file, or a secrets
// manager) is the caller's responsibility; this package does not read them.
func NewSMTPMailer(host string, port int, email, pass string) *SMTPMailer {
	return &SMTPMailer{
		Host:  host,
		Port:  port,
		Email: email,
		Pass:  pass,
	}
}

// Send dispatches an email message through the configured SMTP server.
// The method converts the Message interface to gomail format and sends it using
// the specified SMTP server with STARTTLS encryption (if supported by the server).
// The sender address is taken from the SMTPMailer configuration.
//
// Returns an error if the message could not be sent due to network issues,
// authentication problems, server configuration errors, or invalid message format.
func (s *SMTPMailer) Send(msg Message) error {
	if err := ValidateMessage(msg); err != nil {
		return err
	}
	m := gomail.NewMessage()

	// Set From header using the mailer's configured address with optional display name
	if fromName := msg.GetFromName(); fromName != "" {
		m.SetAddressHeader("From", s.Email, fromName)
	} else {
		m.SetHeader("From", s.Email)
	}

	m.SetHeader("To", msg.GetTo()...)
	m.SetHeader("Subject", msg.GetSubject())
	m.SetBody("text/html", msg.GetBody())

	// Add attachments if any
	for _, attachment := range msg.GetAttachments() {
		m.Attach(attachment.GetFilename(), gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(attachment.GetData())
			return err
		}))
	}

	// Connect to the SMTP server and send the message
	d := gomail.NewDialer(s.Host, s.Port, s.Email, s.Pass)
	return d.DialAndSend(m)
}

// GetFrom returns the SMTP account's email address.
func (s *SMTPMailer) GetFrom() string {
	return s.Email
}
