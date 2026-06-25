package mail

import (
	"errors"
	"fmt"
	"strings"

	netmail "net/mail"
)

// EmailMessage is a concrete implementation of the Message interface.
// It represents an email with recipients, subject, body, and attachments.
// The sender information is handled by the Mailer implementation.
type EmailMessage struct {
	fromName    string       // sender's display name (optional)
	to          []string     // list of recipient email addresses
	subject     string       // email subject line
	body        string       // email body content (typically HTML)
	attachments []Attachment // file attachments
}

// NewMessage creates a new EmailMessage with the specified subject and body.
// The sender information is provided by the Mailer when sending.
// Recipients and attachments can be added using the AddRecipient and AddAttachment methods.
func NewMessage(subject, body string) Message {
	return &EmailMessage{
		subject:     subject,
		body:        body,
		to:          make([]string, 0),
		attachments: make([]Attachment, 0),
	}
}

// GetFromName returns the sender's display name.
func (m *EmailMessage) GetFromName() string {
	return m.fromName
}

// GetTo returns a slice of recipient email addresses.
func (m *EmailMessage) GetTo() []string {
	return append([]string(nil), m.to...)
}

// GetSubject returns the email subject line.
func (m *EmailMessage) GetSubject() string {
	return m.subject
}

// GetBody returns the email body content.
func (m *EmailMessage) GetBody() string {
	return m.body
}

// GetAttachments returns a slice of file attachments.
func (m *EmailMessage) GetAttachments() []Attachment {
	return append([]Attachment(nil), m.attachments...)
}

// AddAttachment adds a file attachment to the email message.
func (m *EmailMessage) AddAttachment(attachment Attachment) {
	if attachment == nil {
		return
	}
	m.attachments = append(m.attachments, attachment)
}

// AddRecipient adds an email address to the list of recipients.
func (m *EmailMessage) AddRecipient(email string) {
	m.to = append(m.to, email)
}

// SetFromName sets the sender's display name.
func (m *EmailMessage) SetFromName(name string) {
	m.fromName = name
}

// ErrInvalidMessage is returned when a message is missing required fields.
var ErrInvalidMessage = errors.New("mail: invalid message")

// ValidateAddress checks that address is a syntactically valid email address.
func ValidateAddress(address string) error {
	if strings.TrimSpace(address) == "" {
		return fmt.Errorf("%w: email address is required", ErrInvalidMessage)
	}
	if containsLineBreak(address) {
		return fmt.Errorf("%w: email address contains a line break", ErrInvalidMessage)
	}
	parsed, err := netmail.ParseAddress(address)
	if err != nil {
		return fmt.Errorf("%w: invalid email address %q", ErrInvalidMessage, address)
	}
	if parsed.Address != address {
		return fmt.Errorf("%w: display names are not accepted in recipient addresses", ErrInvalidMessage)
	}
	return nil
}

// ValidateMessage checks the common requirements before a mailer sends a message.
func ValidateMessage(msg Message) error {
	if msg == nil {
		return fmt.Errorf("%w: message is required", ErrInvalidMessage)
	}
	if strings.TrimSpace(msg.GetSubject()) == "" {
		return fmt.Errorf("%w: subject is required", ErrInvalidMessage)
	}
	if containsLineBreak(msg.GetSubject()) {
		return fmt.Errorf("%w: subject contains a line break", ErrInvalidMessage)
	}
	if containsLineBreak(msg.GetFromName()) {
		return fmt.Errorf("%w: from name contains a line break", ErrInvalidMessage)
	}
	if strings.TrimSpace(msg.GetBody()) == "" {
		return fmt.Errorf("%w: body is required", ErrInvalidMessage)
	}
	recipients := msg.GetTo()
	if len(recipients) == 0 {
		return fmt.Errorf("%w: at least one recipient is required", ErrInvalidMessage)
	}
	for _, recipient := range recipients {
		if err := ValidateAddress(recipient); err != nil {
			return err
		}
	}
	return nil
}

func containsLineBreak(s string) bool {
	return strings.ContainsAny(s, "\r\n")
}
