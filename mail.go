// Package mail provides interfaces and types for sending emails through various providers.
// It defines a common interface for different mail services like Gmail, SMTP, etc.
package mail

// Mailer defines the behavior for any mail provider.
// Implementations should handle the actual sending of messages through their respective services.
type Mailer interface {
	// Send dispatches an email message using the underlying mail service.
	// Returns an error if the message could not be sent.
	Send(msg Message) error

	// GetFrom returns the sender's email address configured for this mailer.
	GetFrom() string
}

// Message interface defines the behavior for email messages.
// It provides methods to access and modify email components like recipients, subject, body, and attachments.
// The sender information is handled by the Mailer implementation.
type Message interface {
	// GetFromName returns the sender's display name (optional).
	GetFromName() string

	// GetTo returns a slice of recipient email addresses.
	GetTo() []string

	// GetSubject returns the email subject line.
	GetSubject() string

	// GetBody returns the email body content (typically HTML).
	GetBody() string

	// GetAttachments returns a slice of file attachments.
	GetAttachments() []Attachment

	// AddAttachment adds a file attachment to the message.
	AddAttachment(attachment Attachment)

	// AddRecipient adds an email address to the list of recipients.
	AddRecipient(email string)

	// SetFromName sets the sender's display name.
	SetFromName(name string)
}

// Attachment interface defines the behavior for email attachments.
// It provides methods to access file data, name, and MIME type information.
type Attachment interface {
	// GetFilename returns the name of the attached file.
	GetFilename() string

	// GetData returns the raw file data as a byte slice.
	GetData() []byte

	// GetMimeType returns the MIME type of the attachment (e.g., "image/png", "application/pdf").
	GetMimeType() string
}
