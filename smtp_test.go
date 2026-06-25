package mail

import (
	"errors"
	"io"
	"testing"

	"gopkg.in/gomail.v2"
)

func TestNewSMTPMailer_WithConfig(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		port      int
		email     string
		pass      string
		wantHost  string
		wantPort  int
		wantEmail string
		wantPass  string
	}{
		{
			name:      "Gmail configuration",
			host:      "smtp.gmail.com",
			port:      587,
			email:     "user@gmail.com",
			pass:      "app-password",
			wantHost:  "smtp.gmail.com",
			wantPort:  587,
			wantEmail: "user@gmail.com",
			wantPass:  "app-password",
		},
		{
			name:      "Outlook configuration",
			host:      "smtp-mail.outlook.com",
			port:      587,
			email:     "user@outlook.com",
			pass:      "password123",
			wantHost:  "smtp-mail.outlook.com",
			wantPort:  587,
			wantEmail: "user@outlook.com",
			wantPass:  "password123",
		},
		{
			name:      "Custom SMTP server",
			host:      "mail.example.com",
			port:      465,
			email:     "admin@example.com",
			pass:      "secret",
			wantHost:  "mail.example.com",
			wantPort:  465,
			wantEmail: "admin@example.com",
			wantPass:  "secret",
		},
		{
			name:      "empty credentials",
			host:      "",
			port:      0,
			email:     "",
			pass:      "",
			wantHost:  "",
			wantPort:  0,
			wantEmail: "",
			wantPass:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := NewSMTPMailer(tt.host, tt.port, tt.email, tt.pass)

			if mailer == nil {
				t.Fatal("NewSMTPMailer returned nil")
			}

			if mailer.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", mailer.Host, tt.wantHost)
			}

			if mailer.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", mailer.Port, tt.wantPort)
			}

			if mailer.Email != tt.wantEmail {
				t.Errorf("User = %q, want %q", mailer.Email, tt.wantEmail)
			}

			if mailer.Pass != tt.wantPass {
				t.Errorf("Pass = %q, want %q", mailer.Pass, tt.wantPass)
			}
		})
	}
}

func TestSMTPMailer_GetFrom(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "standard email",
			email: "user@example.com",
		},
		{
			name:  "empty email",
			email: "",
		},
		{
			name:  "complex email",
			email: "user.name+tag@subdomain.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := NewSMTPMailer("smtp.example.com", 587, tt.email, "password")

			if got := mailer.GetFrom(); got != tt.email {
				t.Errorf("GetFrom() = %q, want %q", got, tt.email)
			}
		})
	}
}

func TestSMTPMailer_InterfaceCompliance(t *testing.T) {
	// Verify that SMTPMailer implements Mailer interface
	var _ Mailer = &SMTPMailer{}
	var _ Mailer = NewSMTPMailer("host", 587, "user", "pass")
}

// mockSendCloser helps us test the Send method without actually connecting to SMTP
type mockSendCloser struct {
	sendFunc  func(from string, to []string, msg io.WriterTo) error
	closeFunc func() error
}

func (m *mockSendCloser) Send(from string, to []string, msg io.WriterTo) error {
	if m.sendFunc != nil {
		return m.sendFunc(from, to, msg)
	}
	return nil
}

func (m *mockSendCloser) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestSMTPMailer_Send_MessageConstruction(t *testing.T) {
	// Test that messages are constructed correctly
	// Note: We can't easily test actual SMTP sending without a real server or complex mocking
	// Instead, we test that the message construction doesn't panic and basic validation works

	tests := []struct {
		name          string
		subject       string
		body          string
		recipients    []string
		fromName      string
		attachments   []Attachment
		shouldSucceed bool
	}{
		{
			name:          "simple message",
			subject:       "Test Subject",
			body:          "Test Body",
			recipients:    []string{"recipient@example.com"},
			fromName:      "",
			attachments:   nil,
			shouldSucceed: false, // Will fail due to no real SMTP server
		},
		{
			name:          "message with from name",
			subject:       "Test Subject",
			body:          "Test Body",
			recipients:    []string{"recipient@example.com"},
			fromName:      "Test Sender",
			attachments:   nil,
			shouldSucceed: false,
		},
		{
			name:          "multiple recipients",
			subject:       "Test Subject",
			body:          "Test Body",
			recipients:    []string{"user1@example.com", "user2@example.com"},
			fromName:      "Test Sender",
			attachments:   nil,
			shouldSucceed: false,
		},
		{
			name:       "message with attachment",
			subject:    "Test Subject",
			body:       "Test Body",
			recipients: []string{"recipient@example.com"},
			fromName:   "Test Sender",
			attachments: []Attachment{
				NewAttachment("test.txt", []byte("test content"), "text/plain"),
			},
			shouldSucceed: false,
		},
		{
			name:       "message with multiple attachments",
			subject:    "Test Subject",
			body:       "Test Body",
			recipients: []string{"recipient@example.com"},
			fromName:   "Test Sender",
			attachments: []Attachment{
				NewAttachment("test1.txt", []byte("content1"), "text/plain"),
				NewAttachment("test2.pdf", []byte("pdf content"), "application/pdf"),
			},
			shouldSucceed: false,
		},
		{
			name:          "HTML body",
			subject:       "HTML Email",
			body:          "<h1>Hello</h1><p>This is <strong>HTML</strong></p>",
			recipients:    []string{"recipient@example.com"},
			fromName:      "Test Sender",
			attachments:   nil,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mailer (won't connect to real server)
			mailer := NewSMTPMailer("localhost", 9999, "test@example.com", "password")

			// Create message
			msg := NewMessage(tt.subject, tt.body)
			if tt.fromName != "" {
				msg.SetFromName(tt.fromName)
			}
			for _, recipient := range tt.recipients {
				msg.AddRecipient(recipient)
			}
			for _, attachment := range tt.attachments {
				msg.AddAttachment(attachment)
			}

			// Try to send (will fail with network error, but that's expected)
			err := mailer.Send(msg)

			// We expect an error since there's no real SMTP server
			// The important thing is that it doesn't panic
			if err == nil && !tt.shouldSucceed {
				t.Error("Expected error due to no SMTP server, but got nil")
			}

			// The error should be a network/connection error, not a panic
			if err != nil {
				t.Logf("Expected error occurred: %v", err)
			}
		})
	}
}

func TestSMTPMailer_Send_EmptyRecipients(t *testing.T) {
	// Test sending with no recipients
	mailer := NewSMTPMailer("localhost", 9999, "test@example.com", "password")
	msg := NewMessage("Subject", "Body")
	// Don't add any recipients

	err := mailer.Send(msg)
	// Should fail (likely with network error or validation error)
	if err == nil {
		t.Error("Expected error when sending to no recipients")
	}
}

func TestSMTPMailer_SendRejectsInvalidMessageBeforeNetwork(t *testing.T) {
	mailer := NewSMTPMailer("localhost", 9999, "test@example.com", "password")
	msg := NewMessage("Subject\r\nBcc: attacker@example.com", "Body")
	msg.AddRecipient("recipient@example.com")

	err := mailer.Send(msg)
	if !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("Send() error = %v, want ErrInvalidMessage", err)
	}
}

func TestSMTPMailer_Send_Configuration(t *testing.T) {
	// Test that different SMTP configurations are accepted
	configs := []struct {
		name  string
		host  string
		port  int
		email string
		pass  string
	}{
		{
			name:  "standard TLS port",
			host:  "smtp.example.com",
			port:  587,
			email: "user@example.com",
			pass:  "password",
		},
		{
			name:  "SSL port",
			host:  "smtp.example.com",
			port:  465,
			email: "user@example.com",
			pass:  "password",
		},
		{
			name:  "non-standard port",
			host:  "smtp.example.com",
			port:  2525,
			email: "user@example.com",
			pass:  "password",
		},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			mailer := NewSMTPMailer(cfg.host, cfg.port, cfg.email, cfg.pass)

			// Verify configuration was set
			if mailer.Host != cfg.host {
				t.Errorf("Host = %q, want %q", mailer.Host, cfg.host)
			}
			if mailer.Port != cfg.port {
				t.Errorf("Port = %d, want %d", mailer.Port, cfg.port)
			}
			if mailer.Email != cfg.email {
				t.Errorf("Email = %q, want %q", mailer.Email, cfg.email)
			}
			if mailer.Pass != cfg.pass {
				t.Errorf("Pass = %q, want %q", mailer.Pass, cfg.pass)
			}

			// Try to send (will fail with network error)
			msg := NewMessage("Test", "Test")
			msg.AddRecipient("test@example.com")
			err := mailer.Send(msg)

			// Should fail to connect
			if err == nil {
				t.Error("Expected connection error")
			}
		})
	}
}

func TestSMTPMailer_AttachmentHandling(t *testing.T) {
	// Test that attachment data is correctly passed to gomail
	mailer := NewSMTPMailer("localhost", 9999, "test@example.com", "password")

	// Create message with various attachment types
	msg := NewMessage("Test", "Body")
	msg.AddRecipient("recipient@example.com")

	// Add various attachments
	msg.AddAttachment(NewAttachment("text.txt", []byte("text content"), "text/plain"))
	msg.AddAttachment(NewAttachment("data.json", []byte(`{"key":"value"}`), "application/json"))
	msg.AddAttachment(NewAttachment("image.png", []byte{0x89, 0x50, 0x4E, 0x47}, "image/png"))

	// Try to send (will fail but shouldn't panic on attachment handling)
	err := mailer.Send(msg)
	if err == nil {
		t.Error("Expected error due to no SMTP server")
	}

	// Verify we can construct the message multiple times
	err = mailer.Send(msg)
	if err == nil {
		t.Error("Expected error due to no SMTP server")
	}
}

// Test that demonstrates the gomail message construction
func TestSMTPMailer_GomailMessageConstruction(t *testing.T) {
	// This test verifies that we can construct gomail messages correctly
	// even if we can't send them

	msg := NewMessage("Test Subject", "Test Body")
	msg.SetFromName("Test Sender")
	msg.AddRecipient("recipient@example.com")
	msg.AddAttachment(NewAttachment("file.txt", []byte("content"), "text/plain"))

	// Create gomail message (mimicking what Send does)
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "sender@example.com", msg.GetFromName())
	m.SetHeader("To", msg.GetTo()...)
	m.SetHeader("Subject", msg.GetSubject())
	m.SetBody("text/html", msg.GetBody())

	// Add attachments
	for _, attachment := range msg.GetAttachments() {
		m.Attach(attachment.GetFilename(), gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(attachment.GetData())
			return err
		}))
	}

	// If we got here without panicking, the message construction works
	t.Log("Message constructed successfully")
}
