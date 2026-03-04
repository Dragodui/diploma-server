package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Dragodui/diploma-server/internal/metrics"
	"gopkg.in/gomail.v2"
)

var (
	// emailRegex validates email format
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidEmailHeader is returned when email header contains forbidden characters
	ErrInvalidEmailHeader = errors.New("email header contains forbidden characters (CRLF injection attempt)")
)

// sanitizeEmailHeader removes CRLF characters to prevent email header injection
// These characters (\r\n) allow attackers to inject arbitrary email headers
func sanitizeEmailHeader(input string) string {
	// Remove all carriage return and line feed characters
	sanitized := strings.ReplaceAll(input, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\n", "")
	// Also remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	return strings.TrimSpace(sanitized)
}

// validateEmail checks if email address is valid and safe
func validateEmail(email string) error {
	// Check for CRLF injection attempts
	if strings.ContainsAny(email, "\r\n\x00") {
		return ErrInvalidEmailHeader
	}

	// Validate email format
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

type Mailer interface {
	Send(to, subject, body string) error
}

type SMTPMailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	// Validate email address to prevent header injection
	if err := validateEmail(to); err != nil {
		return err
	}

	// Sanitize subject to prevent CRLF injection
	// Subject is user-controlled and could contain malicious CRLF sequences
	sanitizedSubject := sanitizeEmailHeader(subject)

	// Sanitize body to prevent CRLF injection in HTML content
	// While body is less critical (it's HTML content, not headers),
	// we still sanitize to prevent potential attacks
	sanitizedBody := sanitizeEmailHeader(body)

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", to) // Already validated above
	msg.SetHeader("Subject", sanitizedSubject)
	msg.SetBody("text/html", sanitizedBody)

	dial := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	start := time.Now()
	err := dial.DialAndSend(msg)
	metrics.EmailSendDuration.Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.EmailsSentTotal.WithLabelValues("smtp", "error").Inc()
		return err
	}
	metrics.EmailsSentTotal.WithLabelValues("smtp", "success").Inc()
	return nil
}

type BrevoMailer struct {
	APIKey string
	From   string
}

func (m *BrevoMailer) Send(to, subject, body string) error {
	start := time.Now()
	defer func() {
		metrics.EmailSendDuration.Observe(time.Since(start).Seconds())
	}()
	payload := map[string]interface{}{
		"sender":      map[string]string{"email": m.From},
		"to":          []map[string]string{{"email": to}},
		"subject":     subject,
		"htmlContent": body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("api-key", m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		metrics.EmailsSentTotal.WithLabelValues("brevo", "error").Inc()
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo API error %d: %s", resp.StatusCode, string(respBody))
	}

	metrics.EmailsSentTotal.WithLabelValues("brevo", "success").Inc()
	return nil
}
