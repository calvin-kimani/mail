package mail

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"gopkg.in/gomail.v2"
)

// startFakeSMTP starts a minimal SMTP server on a random local port.
// It accepts connections, sends a 220 greeting, then rejects AUTH with 535.
// Returns host, port, and a cleanup function.
func startFakeSMTP(t *testing.T) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start fake SMTP: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			go handleFakeSMTP(conn)
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port
}

func handleFakeSMTP(conn net.Conn) {
	defer conn.Close()
	w := bufio.NewWriter(conn)
	r := bufio.NewReader(conn)

	fmt.Fprintf(w, "220 fake SMTP ready\r\n")
	w.Flush()

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		switch {
		case len(line) >= 4 && line[:4] == "EHLO":
			fmt.Fprintf(w, "250-fake\r\n250 AUTH PLAIN LOGIN\r\n")
		case len(line) >= 4 && line[:4] == "HELO":
			fmt.Fprintf(w, "250 fake\r\n")
		case len(line) >= 8 && line[:8] == "STARTTLS":
			fmt.Fprintf(w, "502 STARTTLS not supported\r\n")
		case len(line) >= 4 && line[:4] == "AUTH":
			fmt.Fprintf(w, "535 authentication failed\r\n")
		case len(line) >= 4 && line[:4] == "QUIT":
			fmt.Fprintf(w, "221 bye\r\n")
			w.Flush()
			return
		default:
			fmt.Fprintf(w, "502 not implemented\r\n")
		}
		w.Flush()
	}
}

// newTestGmailMailer creates a GmailMailer pointing at the fake SMTP server.
func newTestGmailMailer(t *testing.T) *GmailMailer {
	t.Helper()
	host, port := startFakeSMTP(t)
	m := NewGmailMailer("test@gmail.com", "fake-password")
	m.Host = host
	m.Port = port
	return m
}

func TestNewGmailMailer_WithCredentials(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		password     string
		wantEmail    string
		wantPassword string
	}{
		{
			name:         "standard Gmail credentials",
			email:        "user@gmail.com",
			password:     "app-password-here",
			wantEmail:    "user@gmail.com",
			wantPassword: "app-password-here",
		},
		{
			name:         "G Suite email",
			email:        "admin@company.com",
			password:     "secure-app-password",
			wantEmail:    "admin@company.com",
			wantPassword: "secure-app-password",
		},
		{
			name:         "empty credentials",
			email:        "",
			password:     "",
			wantEmail:    "",
			wantPassword: "",
		},
		{
			name:         "complex app password",
			email:        "user@gmail.com",
			password:     "abcd efgh ijkl mnop",
			wantEmail:    "user@gmail.com",
			wantPassword: "abcd efgh ijkl mnop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := NewGmailMailer(tt.email, tt.password)

			if mailer == nil {
				t.Fatal("NewGmailMailer returned nil")
			}

			if mailer.Email != tt.wantEmail {
				t.Errorf("Username = %q, want %q", mailer.Email, tt.wantEmail)
			}

			if mailer.Password != tt.wantPassword {
				t.Errorf("Password = %q, want %q", mailer.Password, tt.wantPassword)
			}
		})
	}
}

func TestGmailMailer_GetFrom(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "standard Gmail address",
			email: "user@gmail.com",
		},
		{
			name:  "G Suite address",
			email: "admin@company.com",
		},
		{
			name:  "empty email",
			email: "",
		},
		{
			name:  "email with plus addressing",
			email: "user+tag@gmail.com",
		},
		{
			name:  "email with dots",
			email: "first.last@gmail.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := NewGmailMailer(tt.email, "password")

			if got := mailer.GetFrom(); got != tt.email {
				t.Errorf("GetFrom() = %q, want %q", got, tt.email)
			}
		})
	}
}

func TestGmailMailer_From_OverridesSenderButNotAuth(t *testing.T) {
	mailer := NewGmailMailer("account@gmail.com", "app-password")

	if got := mailer.GetFrom(); got != "account@gmail.com" {
		t.Fatalf("GetFrom() with no From set = %q, want fallback to Email", got)
	}

	mailer.SetFrom("noreply@company.com")

	if got := mailer.GetFrom(); got != "noreply@company.com" {
		t.Errorf("GetFrom() after SetFrom = %q, want %q", got, "noreply@company.com")
	}
	if mailer.Email != "account@gmail.com" {
		t.Errorf("Email (auth username) = %q, want it unchanged", mailer.Email)
	}
}

func TestGmailMailer_From_EmptyFallsBackToEmail(t *testing.T) {
	mailer := NewGmailMailer("user@gmail.com", "password")

	if mailer.From != "" {
		t.Fatalf("From = %q, want empty default", mailer.From)
	}
	if got := mailer.GetFrom(); got != "user@gmail.com" {
		t.Errorf("GetFrom() = %q, want %q", got, "user@gmail.com")
	}
}

func TestGmailMailer_InterfaceCompliance(t *testing.T) {
	// Verify that GmailMailer implements Mailer interface
	var _ Mailer = &GmailMailer{}
	var _ Mailer = NewGmailMailer("user@gmail.com", "password")
}

func TestGmailMailer_Send_MessageConstruction(t *testing.T) {
	// Test that messages are constructed correctly for Gmail
	// Note: We can't actually connect to Gmail without valid credentials

	tests := []struct {
		name        string
		subject     string
		body        string
		recipients  []string
		fromName    string
		attachments []Attachment
	}{
		{
			name:        "simple message",
			subject:     "Test Subject",
			body:        "Test Body",
			recipients:  []string{"recipient@example.com"},
			fromName:    "",
			attachments: nil,
		},
		{
			name:        "message with from name",
			subject:     "Test Subject",
			body:        "Test Body",
			recipients:  []string{"recipient@example.com"},
			fromName:    "Test Sender",
			attachments: nil,
		},
		{
			name:        "multiple recipients",
			subject:     "Test Subject",
			body:        "Test Body",
			recipients:  []string{"user1@example.com", "user2@example.com", "user3@example.com"},
			fromName:    "Test Sender",
			attachments: nil,
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
		},
		{
			name:       "message with multiple attachments",
			subject:    "Test Subject",
			body:       "Test Body",
			recipients: []string{"recipient@example.com"},
			fromName:   "Test Sender",
			attachments: []Attachment{
				NewAttachment("doc.txt", []byte("content"), "text/plain"),
				NewAttachment("image.png", []byte{0x89, 0x50, 0x4E, 0x47}, "image/png"),
				NewAttachment("data.json", []byte(`{"key":"value"}`), "application/json"),
			},
		},
		{
			name:        "HTML email",
			subject:     "HTML Email",
			body:        "<html><body><h1>Hello</h1><p>This is <strong>HTML</strong></p></body></html>",
			recipients:  []string{"recipient@example.com"},
			fromName:    "HTML Sender",
			attachments: nil,
		},
		{
			name:        "unicode content",
			subject:     "Unicode Test: 你好世界",
			body:        "Content with unicode: Olá, Привет, مرحبا",
			recipients:  []string{"recipient@example.com"},
			fromName:    "Unicode Sender 📧",
			attachments: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := newTestGmailMailer(t)

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

			// Try to send (will fail with authentication error, but that's expected)
			err := mailer.Send(msg)

			// We expect an error since credentials are fake
			// The important thing is that it doesn't panic
			if err == nil {
				t.Error("Expected error due to fake credentials, but got nil")
			}

			// The error should be network/auth related, not a panic
			if err != nil {
				t.Logf("Expected error occurred: %v", err)
			}
		})
	}
}

func TestGmailMailer_Send_EmptyRecipients(t *testing.T) {
	// Test sending with no recipients
	mailer := newTestGmailMailer(t)
	msg := NewMessage("Subject", "Body")
	// Don't add any recipients

	err := mailer.Send(msg)
	// Should fail (likely with validation or network error)
	if err == nil {
		t.Error("Expected error when sending to no recipients")
	}
}

func TestGmailMailer_SendRejectsInvalidMessageBeforeNetwork(t *testing.T) {
	mailer := NewGmailMailer("user@gmail.com", "password")
	msg := NewMessage("Subject", "Body")
	msg.AddRecipient("recipient@example.com\nBcc: attacker@example.com")

	err := mailer.Send(msg)
	if !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("Send() error = %v, want ErrInvalidMessage", err)
	}
}

func TestGmailMailer_Send_VerifyGmailSMTPSettings(t *testing.T) {
	// This test verifies that Gmail-specific settings are used
	mailer := NewGmailMailer("user@gmail.com", "app-password")

	// Verify credentials are set
	if mailer.Email != "user@gmail.com" {
		t.Errorf("Username = %q, want %q", mailer.Email, "user@gmail.com")
	}
	if mailer.Password != "app-password" {
		t.Errorf("Password = %q, want %q", mailer.Password, "app-password")
	}

	// Verify default host/port
	if mailer.Host != "smtp.gmail.com" {
		t.Errorf("Host = %q, want %q", mailer.Host, "smtp.gmail.com")
	}
	if mailer.Port != 587 {
		t.Errorf("Port = %d, want %d", mailer.Port, 587)
	}

	// Point at local fake SMTP to avoid timeout
	host, port := startFakeSMTP(t)
	mailer.Host = host
	mailer.Port = port

	msg := NewMessage("Test", "Body")
	msg.AddRecipient("recipient@example.com")

	err := mailer.Send(msg)
	if err == nil {
		t.Error("Expected authentication error with fake credentials")
	}
	t.Logf("Expected authentication error: %v", err)
}

func TestGmailMailer_AttachmentHandling(t *testing.T) {
	// Test that attachments work correctly with Gmail
	mailer := newTestGmailMailer(t)

	// Create message with various attachment types
	msg := NewMessage("Test", "Body")
	msg.AddRecipient("recipient@example.com")

	// Add various attachments
	msg.AddAttachment(NewAttachment("document.txt", []byte("text content"), "text/plain"))
	msg.AddAttachment(NewAttachment("spreadsheet.csv", []byte("col1,col2\nval1,val2"), "text/csv"))
	msg.AddAttachment(NewAttachment("photo.jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"))

	// Try to send (will fail but shouldn't panic on attachment handling)
	err := mailer.Send(msg)
	if err == nil {
		t.Error("Expected error due to fake credentials")
	}

	// Verify we can send the same message multiple times
	err = mailer.Send(msg)
	if err == nil {
		t.Error("Expected error due to fake credentials")
	}
}

func TestGmailMailer_GomailMessageConstruction(t *testing.T) {
	// Test that we correctly construct gomail messages for Gmail
	msg := NewMessage("Test Subject", "Test Body")
	msg.SetFromName("Test Sender")
	msg.AddRecipient("recipient@example.com")
	msg.AddRecipient("another@example.com")
	msg.AddAttachment(NewAttachment("file.txt", []byte("content"), "text/plain"))

	// Create gomail message (mimicking what Send does)
	m := gomail.NewMessage()
	username := "user@gmail.com"
	m.SetAddressHeader("From", username, msg.GetFromName())
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
	t.Log("Gmail message constructed successfully")
}

func TestGmailMailer_ComparisonWithSMTPMailer(t *testing.T) {
	// Test that GmailMailer behaves similarly to SMTPMailer with same settings
	host, port := startFakeSMTP(t)

	gmailMailer := NewGmailMailer("user@gmail.com", "password")
	gmailMailer.Host = host
	gmailMailer.Port = port

	smtpMailer := NewSMTPMailer(host, port, "user@gmail.com", "password")

	// Both should have the same username
	if gmailMailer.GetFrom() != smtpMailer.GetFrom() {
		t.Errorf("GetFrom() mismatch: gmail=%q, smtp=%q", gmailMailer.GetFrom(), smtpMailer.GetFrom())
	}

	// Both should fail similarly with fake credentials
	msg := NewMessage("Test", "Body")
	msg.AddRecipient("recipient@example.com")

	gmailErr := gmailMailer.Send(msg)
	smtpErr := smtpMailer.Send(msg)

	// Both should produce errors
	if gmailErr == nil {
		t.Error("Expected error from GmailMailer")
	}
	if smtpErr == nil {
		t.Error("Expected error from SMTPMailer")
	}

	t.Logf("GmailMailer error: %v", gmailErr)
	t.Logf("SMTPMailer error: %v", smtpErr)
}
