package mail

import (
	"bytes"
	"testing"
)

func TestNewAttachment(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		data     []byte
		mimeType string
	}{
		{
			name:     "text file",
			filename: "document.txt",
			data:     []byte("This is a text document"),
			mimeType: "text/plain",
		},
		{
			name:     "PDF document",
			filename: "report.pdf",
			data:     []byte("%PDF-1.4 fake pdf content"),
			mimeType: "application/pdf",
		},
		{
			name:     "image file",
			filename: "photo.jpg",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG header
			mimeType: "image/jpeg",
		},
		{
			name:     "empty file",
			filename: "empty.txt",
			data:     []byte{},
			mimeType: "text/plain",
		},
		{
			name:     "large file",
			filename: "large.bin",
			data:     make([]byte, 1024*1024), // 1MB
			mimeType: "application/octet-stream",
		},
		{
			name:     "special characters in filename",
			filename: "file with spaces & special-chars.txt",
			data:     []byte("content"),
			mimeType: "text/plain",
		},
		{
			name:     "unicode filename",
			filename: "文档.txt",
			data:     []byte("unicode content"),
			mimeType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := NewAttachment(tt.filename, tt.data, tt.mimeType)

			if attachment == nil {
				t.Fatal("NewAttachment returned nil")
			}

			if attachment.GetFilename() != tt.filename {
				t.Errorf("GetFilename() = %q, want %q", attachment.GetFilename(), tt.filename)
			}

			if attachment.GetMimeType() != tt.mimeType {
				t.Errorf("GetMimeType() = %q, want %q", attachment.GetMimeType(), tt.mimeType)
			}

			gotData := attachment.GetData()
			if !bytes.Equal(gotData, tt.data) {
				t.Errorf("GetData() length = %d, want %d", len(gotData), len(tt.data))
				if len(gotData) == len(tt.data) && len(gotData) < 100 {
					t.Errorf("GetData() = %v, want %v", gotData, tt.data)
				}
			}
		})
	}
}

func TestFileAttachment_GetFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "simple filename",
			filename: "file.txt",
		},
		{
			name:     "long filename",
			filename: "this_is_a_very_long_filename_with_many_characters_that_exceeds_normal_expectations.txt",
		},
		{
			name:     "filename with path separators",
			filename: "folder/subfolder/file.txt",
		},
		{
			name:     "filename with dots",
			filename: "file.backup.2024.01.15.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := NewAttachment(tt.filename, []byte("data"), "text/plain")

			if got := attachment.GetFilename(); got != tt.filename {
				t.Errorf("GetFilename() = %q, want %q", got, tt.filename)
			}
		})
	}
}

func TestFileAttachment_GetData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "simple text",
			data: []byte("Hello, World!"),
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "nil is converted to empty",
			data: nil,
		},
		{
			name: "large data",
			data: bytes.Repeat([]byte("x"), 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := NewAttachment("file.dat", tt.data, "application/octet-stream")

			got := attachment.GetData()
			if tt.data == nil {
				// nil should be stored as-is or converted to empty slice
				if len(got) != 0 {
					t.Errorf("GetData() for nil input = %v, want nil or empty", got)
				}
			} else if !bytes.Equal(got, tt.data) {
				t.Errorf("GetData() mismatch, length = %d, want %d", len(got), len(tt.data))
			}
		})
	}
}

func TestFileAttachment_DataIsCopied(t *testing.T) {
	original := []byte("secret")
	attachment := NewAttachment("secret.txt", original, "text/plain")

	original[0] = 'X'
	if got := string(attachment.GetData()); got != "secret" {
		t.Fatalf("attachment data changed after mutating original slice: %q", got)
	}

	got := attachment.GetData()
	got[1] = 'X'
	if gotAgain := string(attachment.GetData()); gotAgain != "secret" {
		t.Fatalf("attachment data changed after mutating getter result: %q", gotAgain)
	}
}

func TestFileAttachment_GetMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
	}{
		{
			name:     "text/plain",
			mimeType: "text/plain",
		},
		{
			name:     "text/html",
			mimeType: "text/html",
		},
		{
			name:     "application/json",
			mimeType: "application/json",
		},
		{
			name:     "image/png",
			mimeType: "image/png",
		},
		{
			name:     "image/jpeg",
			mimeType: "image/jpeg",
		},
		{
			name:     "image/gif",
			mimeType: "image/gif",
		},
		{
			name:     "application/pdf",
			mimeType: "application/pdf",
		},
		{
			name:     "application/zip",
			mimeType: "application/zip",
		},
		{
			name:     "video/mp4",
			mimeType: "video/mp4",
		},
		{
			name:     "audio/mpeg",
			mimeType: "audio/mpeg",
		},
		{
			name:     "application/octet-stream",
			mimeType: "application/octet-stream",
		},
		{
			name:     "empty mime type",
			mimeType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := NewAttachment("file", []byte("data"), tt.mimeType)

			if got := attachment.GetMimeType(); got != tt.mimeType {
				t.Errorf("GetMimeType() = %q, want %q", got, tt.mimeType)
			}
		})
	}
}

func TestFileAttachment_Immutability(t *testing.T) {
	// Test that modifying the original data doesn't affect the attachment
	originalData := []byte("original content")
	attachment := NewAttachment("file.txt", originalData, "text/plain")

	// Modify the original data
	originalData[0] = 'X'
	originalData[1] = 'X'
	originalData[2] = 'X'

	// Get data from attachment
	attachmentData := attachment.GetData()

	// Verify attachment data is not affected (depends on implementation)
	// If implementation doesn't copy, this test documents that behavior
	t.Logf("Original data after modification: %s", originalData)
	t.Logf("Attachment data: %s", attachmentData)

	// For now, just verify we can retrieve the data
	if len(attachmentData) == 0 {
		t.Error("GetData() returned empty slice")
	}
}

func TestFileAttachment_InterfaceCompliance(t *testing.T) {
	// Verify that FileAttachment implements Attachment interface
	var _ Attachment = &FileAttachment{}
	var _ Attachment = NewAttachment("", nil, "")
}

func TestFileAttachment_MultipleAttachments(t *testing.T) {
	// Test creating multiple attachments with different data types
	attachments := []Attachment{
		NewAttachment("text.txt", []byte("text content"), "text/plain"),
		NewAttachment("data.json", []byte(`{"key":"value"}`), "application/json"),
		NewAttachment("image.png", []byte{0x89, 0x50, 0x4E, 0x47}, "image/png"),
	}

	// Verify each attachment is independent
	for i, att := range attachments {
		if att == nil {
			t.Errorf("Attachment %d is nil", i)
		}
		if att.GetFilename() == "" {
			t.Errorf("Attachment %d has empty filename", i)
		}
		if att.GetMimeType() == "" {
			t.Errorf("Attachment %d has empty mime type", i)
		}
		if len(att.GetData()) == 0 {
			t.Errorf("Attachment %d has empty data", i)
		}
	}

	// Verify they don't interfere with each other
	if attachments[0].GetFilename() == attachments[1].GetFilename() {
		t.Error("Attachments should have different filenames")
	}
}
