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
	Email string // SMTP email/username used for authentication (often, but not always, the sender address)
	Pass  string // SMTP password or app-specific password
	// From is the sender address placed in the From header and SMTP envelope.
	// When empty it falls back to Email. Set this when the auth username is not
	// itself a valid sender address — e.g. Plunk, whose SMTP username is the
	// literal string "plunk" while the sender must be a verified email address.
	From string
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

	// Set From header using the effective sender address with optional display
	// name. The auth username (s.Email) may differ from the sender address, so
	// prefer s.From when set. gomail derives the SMTP envelope MAIL FROM from
	// this header, while authentication below still uses s.Email.
	fromAddr := s.effectiveFrom()
	if fromName := msg.GetFromName(); fromName != "" {
		m.SetAddressHeader("From", fromAddr, fromName)
	} else {
		m.SetHeader("From", fromAddr)
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

// SetFrom sets the sender address used in the From header and SMTP envelope.
// This is independent of the authentication username (Email): set it when the
// auth username is not a valid sender address, e.g. Plunk.
func (s *SMTPMailer) SetFrom(addr string) {
	s.From = addr
}

// GetFrom returns the effective sender address — From when set, otherwise the
// authentication email.
func (s *SMTPMailer) GetFrom() string {
	return s.effectiveFrom()
}

// effectiveFrom resolves the sender address, falling back to the auth email
// when no explicit From is configured.
func (s *SMTPMailer) effectiveFrom() string {
	if s.From != "" {
		return s.From
	}
	return s.Email
}
