package email

import (
	"fmt"
	"html"
	"log/slog"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Service struct {
	config Config
}

func NewFromEnv() *Service {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587
	}

	return &Service{
		config: Config{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     port,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("EMAIL_FROM"),
		},
	}
}

func (s *Service) Send(to, subject, textBody, htmlBody string) error {
	if s.config.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", textBody)
	if htmlBody != "" {
		m.AddAlternative("text/html", htmlBody)
	}

	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("email sent", "to", to, "subject", subject)
	return nil
}

func (s *Service) SendMagicLink(to, token, baseURL string) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)

	subject := "Login to BandCash"
	textBody := fmt.Sprintf(`Hello!

Click the link below to login to BandCash:

%s

This link will expire in 1 hour.

If you didn't request this, you can safely ignore this email.
`, link)

	htmlBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #222; max-width: 560px;">
  <h2 style="margin: 0 0 12px;">Login to BandCash</h2>
  <p style="margin: 0 0 12px;">Click the button below to sign in:</p>
  <p style="margin: 0 0 16px;">
    <a href="%s" style="display: inline-block; background: #1f6feb; color: #fff; text-decoration: none; padding: 10px 14px; border-radius: 6px;">Sign in to BandCash</a>
  </p>
  <p style="margin: 0 0 8px;">Or copy and paste this link into your browser:</p>
  <p style="margin: 0 0 16px;"><a href="%s">%s</a></p>
  <p style="margin: 0 0 8px;">This link will expire in 1 hour.</p>
  <p style="margin: 0; color: #555;">If you didn't request this, you can safely ignore this email.</p>
</div>
`, link, link, link)

	return s.Send(to, subject, textBody, htmlBody)
}

func (s *Service) SendGroupInvitation(to, groupName, token, baseURL string) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)

	subject := fmt.Sprintf("You've been invited to %s on BandCash", groupName)
	escapedGroupName := html.EscapeString(groupName)
	textBody := fmt.Sprintf(`Hello!

You've been invited to join "%s" on BandCash.

Click the link below to accept the invitation:

%s

This link will expire in 1 hour.

If you didn't expect this invitation, you can safely ignore this email.
`, groupName, link)

	htmlBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #222; max-width: 560px;">
  <h2 style="margin: 0 0 12px;">You're invited to BandCash</h2>
  <p style="margin: 0 0 12px;">You've been invited to join <strong>%s</strong>.</p>
  <p style="margin: 0 0 16px;">
    <a href="%s" style="display: inline-block; background: #1f6feb; color: #fff; text-decoration: none; padding: 10px 14px; border-radius: 6px;">Accept invitation</a>
  </p>
  <p style="margin: 0 0 8px;">Or copy and paste this link into your browser:</p>
  <p style="margin: 0 0 16px;"><a href="%s">%s</a></p>
  <p style="margin: 0 0 8px;">This link will expire in 1 hour.</p>
  <p style="margin: 0; color: #555;">If you didn't expect this invitation, you can safely ignore this email.</p>
</div>
`, escapedGroupName, link, link, link)

	return s.Send(to, subject, textBody, htmlBody)
}
