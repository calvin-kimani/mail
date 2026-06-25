package mail

import (
	"errors"
	"testing"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		body    string
	}{
		{
			name:    "basic message",
			subject: "Test Subject",
			body:    "Test Body",
		},
		{
			name:    "empty subject and body",
			subject: "",
			body:    "",
		},
		{
			name:    "HTML body",
			subject: "HTML Email",
			body:    "<h1>Hello</h1><p>This is HTML</p>",
		},
		{
			name:    "long subject",
			subject: "This is a very long subject line that exceeds normal length expectations for an email subject",
			body:    "Short body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(tt.subject, tt.body)

			if msg == nil {
				t.Fatal("NewMessage returned nil")
			}

			if msg.GetSubject() != tt.subject {
				t.Errorf("GetSubject() = %q, want %q", msg.GetSubject(), tt.subject)
			}

			if msg.GetBody() != tt.body {
				t.Errorf("GetBody() = %q, want %q", msg.GetBody(), tt.body)
			}

			// Verify default values
			if msg.GetFromName() != "" {
				t.Errorf("GetFromName() = %q, want empty string", msg.GetFromName())
			}

			if len(msg.GetTo()) != 0 {
				t.Errorf("GetTo() length = %d, want 0", len(msg.GetTo()))
			}

			if len(msg.GetAttachments()) != 0 {
				t.Errorf("GetAttachments() length = %d, want 0", len(msg.GetAttachments()))
			}
		})
	}
}

func TestEmailMessage_AddRecipient(t *testing.T) {
	tests := []struct {
		name       string
		recipients []string
	}{
		{
			name:       "single recipient",
			recipients: []string{"user@example.com"},
		},
		{
			name:       "multiple recipients",
			recipients: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:       "duplicate recipients",
			recipients: []string{"user@example.com", "user@example.com"},
		},
		{
			name:       "empty string recipient",
			recipients: []string{""},
		},
		{
			name:       "invalid email format",
			recipients: []string{"not-an-email", "also@invalid", "valid@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage("Subject", "Body")

			for _, recipient := range tt.recipients {
				msg.AddRecipient(recipient)
			}

			recipients := msg.GetTo()
			if len(recipients) != len(tt.recipients) {
				t.Errorf("GetTo() length = %d, want %d", len(recipients), len(tt.recipients))
			}

			for i, want := range tt.recipients {
				if i >= len(recipients) {
					t.Errorf("Missing recipient at index %d", i)
					continue
				}
				if recipients[i] != want {
					t.Errorf("GetTo()[%d] = %q, want %q", i, recipients[i], want)
				}
			}
		})
	}
}

func TestEmailMessage_SetFromName(t *testing.T) {
	tests := []struct {
		name     string
		fromName string
	}{
		{
			name:     "simple name",
			fromName: "John Doe",
		},
		{
			name:     "empty name",
			fromName: "",
		},
		{
			name:     "company name",
			fromName: "DotteGigs Support Team",
		},
		{
			name:     "name with special characters",
			fromName: "João Silva (Support)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage("Subject", "Body")
			msg.SetFromName(tt.fromName)

			if msg.GetFromName() != tt.fromName {
				t.Errorf("GetFromName() = %q, want %q", msg.GetFromName(), tt.fromName)
			}
		})
	}
}

func TestEmailMessage_AddAttachment(t *testing.T) {
	tests := []struct {
		name        string
		attachments []Attachment
	}{
		{
			name: "single attachment",
			attachments: []Attachment{
				NewAttachment("file.txt", []byte("content"), "text/plain"),
			},
		},
		{
			name: "multiple attachments",
			attachments: []Attachment{
				NewAttachment("file1.txt", []byte("content1"), "text/plain"),
				NewAttachment("file2.pdf", []byte("pdf content"), "application/pdf"),
				NewAttachment("image.png", []byte("png data"), "image/png"),
			},
		},
		{
			name: "empty attachment data",
			attachments: []Attachment{
				NewAttachment("empty.txt", []byte{}, "text/plain"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage("Subject", "Body")

			for _, attachment := range tt.attachments {
				msg.AddAttachment(attachment)
			}

			attachments := msg.GetAttachments()
			if len(attachments) != len(tt.attachments) {
				t.Errorf("GetAttachments() length = %d, want %d", len(attachments), len(tt.attachments))
			}

			for i, want := range tt.attachments {
				if i >= len(attachments) {
					t.Errorf("Missing attachment at index %d", i)
					continue
				}
				got := attachments[i]
				if got.GetFilename() != want.GetFilename() {
					t.Errorf("Attachment[%d] filename = %q, want %q", i, got.GetFilename(), want.GetFilename())
				}
				if got.GetMimeType() != want.GetMimeType() {
					t.Errorf("Attachment[%d] mime type = %q, want %q", i, got.GetMimeType(), want.GetMimeType())
				}
			}
		})
	}
}

func TestEmailMessage_GettersReturnCopies(t *testing.T) {
	msg := NewMessage("Subject", "Body")
	msg.AddRecipient("user@example.com")
	msg.AddAttachment(NewAttachment("file.txt", []byte("content"), "text/plain"))

	recipients := msg.GetTo()
	recipients[0] = "mutated@example.com"
	if got := msg.GetTo()[0]; got != "user@example.com" {
		t.Fatalf("recipient mutated through getter: %q", got)
	}

	attachments := msg.GetAttachments()
	attachments[0] = NewAttachment("other.txt", []byte("content"), "text/plain")
	if got := msg.GetAttachments()[0].GetFilename(); got != "file.txt" {
		t.Fatalf("attachment mutated through getter: %q", got)
	}
}

func TestValidateAddress(t *testing.T) {
	if err := ValidateAddress("user@example.com"); err != nil {
		t.Fatalf("ValidateAddress() error = %v", err)
	}

	for _, address := range []string{"", "not-an-email", "User <user@example.com>"} {
		if err := ValidateAddress(address); !errors.Is(err, ErrInvalidMessage) {
			t.Fatalf("ValidateAddress(%q) error = %v, want ErrInvalidMessage", address, err)
		}
	}
}

func TestValidateMessage(t *testing.T) {
	msg := NewMessage("Subject", "<p>Body</p>")
	msg.AddRecipient("user@example.com")
	if err := ValidateMessage(msg); err != nil {
		t.Fatalf("ValidateMessage() error = %v", err)
	}

	tests := []struct {
		name  string
		build func() Message
	}{
		{name: "nil message", build: func() Message { return nil }},
		{name: "missing subject", build: func() Message {
			m := NewMessage("", "Body")
			m.AddRecipient("user@example.com")
			return m
		}},
		{name: "missing body", build: func() Message {
			m := NewMessage("Subject", "")
			m.AddRecipient("user@example.com")
			return m
		}},
		{name: "missing recipient", build: func() Message {
			return NewMessage("Subject", "Body")
		}},
		{name: "invalid recipient", build: func() Message {
			m := NewMessage("Subject", "Body")
			m.AddRecipient("bad")
			return m
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateMessage(tt.build()); !errors.Is(err, ErrInvalidMessage) {
				t.Fatalf("ValidateMessage() error = %v, want ErrInvalidMessage", err)
			}
		})
	}
}

func TestValidateMessageRejectsHeaderInjection(t *testing.T) {
	tests := []struct {
		name  string
		build func() Message
	}{
		{name: "subject CRLF", build: func() Message {
			m := NewMessage("Welcome\r\nBcc: attacker@example.com", "Body")
			m.AddRecipient("user@example.com")
			return m
		}},
		{name: "from name CRLF", build: func() Message {
			m := NewMessage("Welcome", "Body")
			m.SetFromName("Support\nReply-To: attacker@example.com")
			m.AddRecipient("user@example.com")
			return m
		}},
		{name: "recipient CRLF", build: func() Message {
			m := NewMessage("Welcome", "Body")
			m.AddRecipient("user@example.com\r\nBcc: attacker@example.com")
			return m
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateMessage(tt.build()); !errors.Is(err, ErrInvalidMessage) {
				t.Fatalf("ValidateMessage() error = %v, want ErrInvalidMessage", err)
			}
		})
	}
}

func TestEmailMessage_CompleteWorkflow(t *testing.T) {
	// Test a complete workflow with all features
	msg := NewMessage("Test Email", "<h1>Hello World</h1>")

	// Set sender name
	msg.SetFromName("Test Sender")
	if msg.GetFromName() != "Test Sender" {
		t.Errorf("GetFromName() = %q, want %q", msg.GetFromName(), "Test Sender")
	}

	// Add recipients
	msg.AddRecipient("user1@example.com")
	msg.AddRecipient("user2@example.com")
	recipients := msg.GetTo()
	if len(recipients) != 2 {
		t.Fatalf("GetTo() length = %d, want 2", len(recipients))
	}

	// Add attachments
	attachment1 := NewAttachment("doc.pdf", []byte("pdf content"), "application/pdf")
	attachment2 := NewAttachment("image.jpg", []byte("jpg data"), "image/jpeg")
	msg.AddAttachment(attachment1)
	msg.AddAttachment(attachment2)
	attachments := msg.GetAttachments()
	if len(attachments) != 2 {
		t.Fatalf("GetAttachments() length = %d, want 2", len(attachments))
	}

	// Verify all fields
	if msg.GetSubject() != "Test Email" {
		t.Errorf("GetSubject() = %q, want %q", msg.GetSubject(), "Test Email")
	}
	if msg.GetBody() != "<h1>Hello World</h1>" {
		t.Errorf("GetBody() = %q, want %q", msg.GetBody(), "<h1>Hello World</h1>")
	}
}

func TestEmailMessage_InterfaceCompliance(t *testing.T) {
	// Verify that EmailMessage implements Message interface
	var _ Message = &EmailMessage{}
	var _ Message = NewMessage("", "")
}
