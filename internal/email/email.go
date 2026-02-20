package email

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	ctxi18n "github.com/invopop/ctxi18n"
	ctxi18ncore "github.com/invopop/ctxi18n/i18n"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"

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

var (
	serviceOnce  sync.Once
	serviceInst  *Service
	i18nLoadOnce sync.Once
)

func ensureI18nLoaded() {
	i18nLoadOnce.Do(func() {
		if err := appi18n.Load(); err != nil {
			slog.Error("email: failed to load i18n", "err", err)
		}
	})
}

func emailContext(ctx context.Context) context.Context {
	ensureI18nLoaded()
	if ctx == nil {
		ctx = context.Background()
	}
	if ctxi18n.Locale(ctx) != nil {
		return ctx
	}
	withLocale, err := ctxi18n.WithLocale(ctx, appi18n.DefaultLocale)
	if err != nil {
		return ctx
	}
	return withLocale
}

func NewFromEnv() *Service {
	env := utils.Env()

	return &Service{
		config: Config{
			Host:     env.SMTPHost,
			Port:     env.SMTPPort,
			Username: env.SMTPUser,
			Password: env.SMTPPass,
			From:     env.EmailFrom,
		},
	}
}

func Email() *Service {
	serviceOnce.Do(func() {
		serviceInst = NewFromEnv()
	})
	return serviceInst
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

func (s *Service) SendMagicLink(ctx context.Context, to, token, baseURL string) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)
	ctx = emailContext(ctx)

	subject := ctxi18ncore.T(ctx, "email.magic_link.subject")
	textBody, err := utils.RenderComponentString(ctx, MagicLinkText(link))
	if err != nil {
		return fmt.Errorf("failed to render magic-link text template: %w", err)
	}
	htmlBody, err := utils.RenderComponentString(ctx, MagicLinkHTML(link))
	if err != nil {
		return fmt.Errorf("failed to render magic-link HTML template: %w", err)
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) SendGroupInvitation(ctx context.Context, to, groupName, token, baseURL string) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)
	ctx = emailContext(ctx)

	subject := ctxi18ncore.T(ctx, "email.invite.subject", groupName)
	textBody, err := utils.RenderComponentString(ctx, GroupInvitationText(groupName, link))
	if err != nil {
		return fmt.Errorf("failed to render invite text template: %w", err)
	}
	htmlBody, err := utils.RenderComponentString(ctx, GroupInvitationHTML(groupName, link))
	if err != nil {
		return fmt.Errorf("failed to render invite HTML template: %w", err)
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}
