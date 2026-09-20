package mail

import (
	"io"

	"gopkg.in/gomail.v2"
)

// Ensure GmailMailer implements the Mailer interface at compile time.
var _ Mailer = &GmailMailer{}

// GmailMailer is a mail service implementation that sends emails through Gmail's SMTP server.
// It uses the standard SMTP protocol with Gmail's servers (smtp.gmail.com:587).
type GmailMailer struct {
	Email    string // Gmail email address (used for authentication and, by default, as the sender)
	Password string // Gmail app password (not regular password)
	Host     string // SMTP host (default: smtp.gmail.com)
	Port     int    // SMTP port (default: 587)
	// From overrides the sender address in the From header and SMTP envelope.
	// When empty it falls back to Email — the normal case for Gmail, where the
	// account address is also the sender.
	From string
}

// NewGmailMailer creates a new Gmail mailer for the given account.
//
// For Gmail you must use an "App Password" rather than your regular account
// password; app passwords can be generated in your Google Account security
// settings. Host and Port default to smtp.gmail.com:587 and may be overridden
// on the returned struct.
//
// Resolving these credentials (from the environment, a config file, or a
// secrets manager) is the caller's responsibility; this package does not read
// them.
func NewGmailMailer(email, password string) *GmailMailer {
	return &GmailMailer{
		Email:    email,
		Password: password,
		Host:     "smtp.gmail.com",
		Port:     587,
	}
}

// Send dispatches an email message through Gmail's SMTP server.
// The method converts the Message interface to gomail format and sends it using
// Gmail's SMTP server (smtp.gmail.com) on port 587 with TLS encryption.
// The sender address is taken from the GmailMailer configuration.
//
// Returns an error if the message could not be sent due to network issues,
// authentication problems, or invalid message format.
func (g *GmailMailer) Send(msg Message) error {
	if err := ValidateMessage(msg); err != nil {
		return err
	}
	m := gomail.NewMessage()

	// Set From header using the effective sender address with optional display
	// name. Authentication below still uses g.Email.
	fromAddr := g.effectiveFrom()
	if fromName := msg.GetFromName(); fromName != "" {
		m.SetAddressHeader("From", fromAddr, fromName)
	} else {
		m.SetHeader("From", fromAddr)
	}

	m.SetHeader("To", msg.GetTo()...)
	m.SetHeader("Subject", msg.GetSubject())
	m.SetBody("text/html", msg.GetBody())

	// Add attachments if any
	for _, a := range msg.GetAttachments() {
		m.Attach(a.GetFilename(), gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(a.GetData())
			return err
		}))
	}

	// Connect to the SMTP server and send the message
	host := g.Host
	if host == "" {
		host = "smtp.gmail.com"
	}
	port := g.Port
	if port == 0 {
		port = 587
	}
	d := gomail.NewDialer(host, port, g.Email, g.Password)
	return d.DialAndSend(m)
}

// SetFrom sets the sender address used in the From header and SMTP envelope,
// independent of the authentication email (Email).
func (g *GmailMailer) SetFrom(addr string) {
	g.From = addr
}

// GetFrom returns the effective sender address — From when set, otherwise the
// Gmail account's email address.
func (g *GmailMailer) GetFrom() string {
	return g.effectiveFrom()
}

// effectiveFrom resolves the sender address, falling back to the account email
// when no explicit From is configured.
func (g *GmailMailer) effectiveFrom() string {
	if g.From != "" {
		return g.From
	}
	return g.Email
}
