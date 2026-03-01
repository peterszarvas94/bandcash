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

func contextWithLocale(base context.Context, locale string) context.Context {
	withLocale, err := ctxi18n.WithLocale(base, locale)
	if err != nil {
		return base
	}
	return withLocale
}

func verifyLink(baseURL, token, locale string) string {
	return fmt.Sprintf("%s/auth/verify?token=%s&lang=%s", baseURL, token, locale)
}

func joinBilingualText(huText, enText string) string {
	return strings.TrimSpace("Find english below.\n\n" + strings.TrimSpace(huText) + "\n\n---\n\n" + strings.TrimSpace(enText))
}

func joinBilingualHTML(huHTML, enHTML string) string {
	return strings.TrimSpace(`<div style="margin:0;padding:0 0 12px;font-family:system-ui,-apple-system,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;color:#1a1a1a;">Find english below.</div>` + strings.TrimSpace(huHTML) + `<div style="height:12px"></div>` + strings.TrimSpace(enHTML))
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
	htmlBody, textBody, subject, err := s.buildMagicLinkBodies(ctx, token, baseURL)
	if err != nil {
		return err
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) PreviewMagicLinkHTML(ctx context.Context, token, baseURL string) (string, error) {
	htmlBody, _, _, err := s.buildMagicLinkBodies(ctx, token, baseURL)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(htmlBody), nil
}

func (s *Service) buildMagicLinkBodies(ctx context.Context, token, baseURL string) (string, string, string, error) {
	ctx = emailContext(ctx)
	huCtx := contextWithLocale(ctx, "hu")
	enCtx := contextWithLocale(ctx, "en")

	huLink := verifyLink(baseURL, token, "hu")
	enLink := verifyLink(baseURL, token, "en")

	subject := ctxi18ncore.T(enCtx, "email.magic_link.subject")
	huTextBody, err := utils.RenderHTML(huCtx, MagicLinkText(huLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render magic-link text template: %w", err)
	}
	enTextBody, err := utils.RenderHTML(enCtx, MagicLinkText(enLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render magic-link text template: %w", err)
	}
	huHTMLBody, err := utils.RenderHTML(huCtx, MagicLinkHTML(huLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render magic-link HTML template: %w", err)
	}
	enHTMLBody, err := utils.RenderHTML(enCtx, MagicLinkHTML(enLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render magic-link HTML template: %w", err)
	}

	textBody := joinBilingualText(huTextBody, enTextBody)
	htmlBody := joinBilingualHTML(huHTMLBody, enHTMLBody)
	return htmlBody, textBody, subject, nil
}

func (s *Service) SendGroupInvitation(ctx context.Context, to, groupName, token, baseURL string) error {
	htmlBody, textBody, subject, err := s.buildGroupInvitationBodies(ctx, groupName, token, baseURL)
	if err != nil {
		return err
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) PreviewGroupInvitationHTML(ctx context.Context, groupName, token, baseURL string) (string, error) {
	htmlBody, _, _, err := s.buildGroupInvitationBodies(ctx, groupName, token, baseURL)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(htmlBody), nil
}

func (s *Service) buildGroupInvitationBodies(ctx context.Context, groupName, token, baseURL string) (string, string, string, error) {
	ctx = emailContext(ctx)
	huCtx := contextWithLocale(ctx, "hu")
	enCtx := contextWithLocale(ctx, "en")

	huLink := verifyLink(baseURL, token, "hu")
	enLink := verifyLink(baseURL, token, "en")

	subject := ctxi18ncore.T(enCtx, "email.invite.subject", groupName)
	huTextBody, err := utils.RenderHTML(huCtx, GroupInvitationText(groupName, huLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite text template: %w", err)
	}
	enTextBody, err := utils.RenderHTML(enCtx, GroupInvitationText(groupName, enLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite text template: %w", err)
	}
	huHTMLBody, err := utils.RenderHTML(huCtx, GroupInvitationHTML(groupName, huLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite HTML template: %w", err)
	}
	enHTMLBody, err := utils.RenderHTML(enCtx, GroupInvitationHTML(groupName, enLink))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite HTML template: %w", err)
	}

	textBody := joinBilingualText(huTextBody, enTextBody)
	htmlBody := joinBilingualHTML(huHTMLBody, enHTMLBody)
	return htmlBody, textBody, subject, nil
}
