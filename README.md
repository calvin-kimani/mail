# Mail

Go interfaces and implementations for sending email.

`github.com/calvin-kimani/mail` provides provider-neutral email primitives plus SMTP and Gmail SMTP implementations. It is intended for applications that want a simple interface for composing HTML email messages, rendering reusable templates, attaching files, validating recipients, and sending through SMTP-compatible providers.

Application-specific templates, branding, password reset URLs, verification URLs, notification copy, and business workflows belong in the application using this module.

## Installation

```sh
go get github.com/calvin-kimani/mail
```

```go
import "github.com/calvin-kimani/mail"
```

## Features

- Common `Mailer`, `Message`, and `Attachment` interfaces.
- HTML email message implementation.
- File attachment implementation.
- Generic SMTP mailer.
- Gmail SMTP mailer.
- HTML template rendering with `html/template`.
- Template-aware sending through `TemplateMailer`.
- Recipient address validation.
- Message validation before sending.
- Header-injection checks for subjects, sender names, and recipients.
- Defensive copying for attachment data.

## Quick Start

```go
mailer := mail.NewSMTPMailer(
    "smtp.example.com",
    587,
    "sender@example.com",
    os.Getenv("SMTP_PASSWORD"),
)

msg := mail.NewMessage("Welcome", "<p>Hello.</p>")
msg.AddRecipient("customer@example.com")
msg.SetFromName("Support")

if err := mail.ValidateMessage(msg); err != nil {
    return err
}

return mailer.Send(msg)
```

## Interfaces

### Mailer

```go
type Mailer interface {
    Send(msg Message) error
    GetFrom() string
}
```

### Message

```go
type Message interface {
    GetFromName() string
    GetTo() []string
    GetSubject() string
    GetBody() string
    GetAttachments() []Attachment
    AddAttachment(attachment Attachment)
    AddRecipient(email string)
    SetFromName(name string)
}
```

### Attachment

```go
type Attachment interface {
    GetFilename() string
    GetData() []byte
    GetMimeType() string
}
```

## SMTP

```go
mailer := mail.NewSMTPMailer(
    "smtp.office365.com",
    587,
    "sender@example.com",
    os.Getenv("SMTP_PASSWORD"),
)
```

`NewSMTPMailer` takes its settings as explicit arguments. Resolving them — from the environment, a config file, or a secrets manager — is the caller's responsibility; this module does not read configuration on its own.

## Gmail

```go
mailer := mail.NewGmailMailer(
    "sender@gmail.com",
    os.Getenv("GMAIL_APP_PASSWORD"),
)
```

`NewGmailMailer` takes the account address and app password as explicit arguments; use an app password rather than the account's normal password. Host and port default to `smtp.gmail.com:587` and can be overridden on the returned mailer.

## Attachments

```go
attachment := mail.NewAttachment("receipt.pdf", pdfBytes, "application/pdf")
msg.AddAttachment(attachment)
```

Common MIME types:

- `application/pdf`
- `text/plain`
- `image/png`
- `image/jpeg`
- `application/zip`

Nil attachments are ignored by `AddAttachment`.

## Validation

Validate a recipient address:

```go
if err := mail.ValidateAddress("customer@example.com"); err != nil {
    return err
}
```

Validate a message:

```go
if err := mail.ValidateMessage(msg); err != nil {
    return err
}
```

`ValidateMessage` checks:

- message is present
- subject is present
- subject does not contain CRLF line breaks
- body is present
- at least one recipient is present
- all recipients are syntactically valid
- sender display name does not contain CRLF line breaks

Display-name formatted recipient addresses such as `User <user@example.com>` are rejected. Pass recipient addresses as plain email addresses and use `SetFromName` for the sender display name.

`SMTPMailer.Send`, `GmailMailer.Send`, and `TemplateMailer.SendTemplate` validate messages before sending.

## Templates

The module includes reusable HTML template rendering built on Go's `html/template`.

### Render a Template Body

```go
renderer, err := mail.NewTemplateRenderer(
    "templates/base.html",
    "templates/password_reset.html",
)
if err != nil {
    return err
}

body, err := renderer.Render("email", map[string]any{
    "Name":     "Ada",
    "ResetURL": "https://example.com/reset?token=abc",
})
if err != nil {
    return err
}

msg := mail.NewMessage("Password reset", body)
msg.AddRecipient("ada@example.com")
```

### Render from an Embedded Filesystem

```go
//go:embed templates/*.html
var templates embed.FS

renderer, err := mail.NewTemplateRendererFS(templates, "templates/*.html")
if err != nil {
    return err
}

msg, err := renderer.Message("Welcome", "welcome", map[string]string{
    "Name": "Ada",
})
if err != nil {
    return err
}
msg.AddRecipient("ada@example.com")
```

### Send a Template Email

`TemplateMailer` wraps any `Mailer`, renders the selected template, validates the resulting message, and sends it.

```go
renderer, err := mail.NewTemplateRendererFS(templates, "templates/*.html")
if err != nil {
    return err
}

templateMailer, err := mail.NewTemplateMailer(mailer, renderer)
if err != nil {
    return err
}

return templateMailer.SendTemplate(mail.TemplateMessage{
    Template: "password_reset",
    Subject:  "Reset your password",
    To:       []string{"ada@example.com"},
    FromName: "Support",
    Data: map[string]any{
        "ResetURL": "https://example.com/reset?token=abc",
    },
})
```

### Layouts and Partials

Use normal `html/template` definitions for layouts and partials:

```html
{{define "email"}}
<!doctype html>
<html>
  <body>
    {{template "content" .}}
  </body>
</html>
{{end}}
```

```html
{{define "content"}}
<h1>Hello {{.Name}}</h1>
{{end}}
```

Render the layout template name:

```go
body, err := renderer.Render("email", data)
```

Application-specific templates, branding, URLs, and copy remain application-owned. This module provides the parsing, rendering, validation, and send orchestration.

Templates are rendered with Go's `html/template`, which contextually escapes untrusted data in HTML text and attributes. Do not cast untrusted user input to trusted template types such as `template.HTML` or `template.URL`.

## Security Notes

- Use `ValidateMessage` before passing messages to custom mailers.
- Subjects, sender display names, and recipient addresses containing `\r` or `\n` are rejected to prevent header injection.
- Recipient addresses must be plain email addresses, not display-name formatted headers.
- Attachment data is copied when created and when read, so callers cannot mutate an attachment through retained slices.
- Template rendering uses `html/template` escaping, but application code is still responsible for only marking trusted content as safe HTML or safe URLs.

## Errors

Validation helpers return errors wrapping `ErrInvalidMessage`.

Template helpers return errors wrapping `ErrTemplate`.

SMTP and Gmail send operations return errors from the underlying SMTP dial/send operation.

## Testing

```sh
go test ./...
```

Unit tests cover message behavior, attachments, validation, template rendering, template sending, and mailer construction. They do not require sending live email.
