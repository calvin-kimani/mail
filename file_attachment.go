package mail

// FileAttachment is a concrete implementation of the Attachment interface.
// It represents a file attachment with its content, filename, and MIME type.
type FileAttachment struct {
	filename string // original filename of the attachment
	data     []byte // raw file content as bytes
	mimeType string // MIME type (e.g., "image/png", "application/pdf", "text/plain")
}

// NewAttachment creates a new file attachment with the specified filename, data, and MIME type.
//
// Parameters:
//   - filename: The name of the file as it should appear in the email
//   - data: The raw file content as a byte slice
//   - mimeType: The MIME type of the file (e.g., "image/jpeg", "application/pdf")
//
// Common MIME types:
//   - Images: "image/jpeg", "image/png", "image/gif"
//   - Documents: "application/pdf", "text/plain", "application/msword"
//   - Archives: "application/zip", "application/x-tar"
func NewAttachment(filename string, data []byte, mimeType string) Attachment {
	copied := append([]byte(nil), data...)
	return &FileAttachment{
		filename: filename,
		data:     copied,
		mimeType: mimeType,
	}
}

// GetFilename returns the filename of the attachment as it will appear in the email.
func (a *FileAttachment) GetFilename() string {
	return a.filename
}

// GetData returns the raw file content as a byte slice.
func (a *FileAttachment) GetData() []byte {
	return append([]byte(nil), a.data...)
}

// GetMimeType returns the MIME type of the attachment.
// This helps email clients properly handle and display the attachment.
func (a *FileAttachment) GetMimeType() string {
	return a.mimeType
}
