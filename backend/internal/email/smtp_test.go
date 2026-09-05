package email

import (
	"testing"
	"time"

	"fejd-backend/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNewSender(t *testing.T) {
	s := NewSender(config.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		SMTPUser: "user",
		SMTPPass: "pass",
		SMTPFrom: "no-reply@example.com",
	})

	assert.Equal(t, "smtp.example.com", s.host)
	assert.Equal(t, 587, s.port)
	assert.Equal(t, "user", s.user)
	assert.Equal(t, "pass", s.pass)
	assert.Equal(t, "no-reply@example.com", s.from)
}

func TestBuildMessage(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	msg := string(buildMessage("no-reply@fejd.local", "owner@example.com", "Hello", "Line1\nLine2", now))

	assert.Contains(t, msg, "From: no-reply@fejd.local\r\n")
	assert.Contains(t, msg, "To: owner@example.com\r\n")
	assert.Contains(t, msg, "Subject: Hello\r\n")
	assert.Contains(t, msg, "Date: Tue, 02 Jan 2024 03:04:05 +0000\r\n")
	assert.Contains(t, msg, "MIME-Version: 1.0\r\n")
	assert.Contains(t, msg, "Content-Type: text/plain; charset=\"utf-8\"\r\n")
	assert.Contains(t, msg, "\r\n\r\n")
	assert.Contains(t, msg, "Line1\nLine2\r\n")
}
