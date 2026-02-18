package email

import (
	"fmt"
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

func (s *Service) Send(to, subject, body string) error {
	if s.config.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

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
	body := fmt.Sprintf(`Hello!

Click the link below to login to BandCash:

%s

This link will expire in 1 hour.

If you didn't request this, you can safely ignore this email.
`, link)

	return s.Send(to, subject, body)
}

func (s *Service) SendGroupInvitation(to, groupName, token, baseURL string) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)

	subject := fmt.Sprintf("You've been invited to %s on BandCash", groupName)
	body := fmt.Sprintf(`Hello!

You've been invited to join "%s" on BandCash.

Click the link below to accept the invitation:

%s

This link will expire in 1 hour.

If you didn't expect this invitation, you can safely ignore this email.
`, groupName, link)

	return s.Send(to, subject, body)
}
