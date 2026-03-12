package email

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
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

func groupLink(baseURL, groupID string) string {
	return fmt.Sprintf("%s/groups/%s", baseURL, groupID)
}

func dashboardLink(baseURL string) string {
	return fmt.Sprintf("%s/dashboard", baseURL)
}

func joinBilingualText(huText, enText string) string {
	return strings.TrimSpace("English follows below.\n\n" + strings.TrimSpace(huText) + "\n\n---\n\n" + strings.TrimSpace(enText))
}

func joinBilingualHTML(ctx context.Context, huHTML, enHTML string) (string, error) {
	rendered, err := utils.RenderHTML(ctx, BilingualEmailHTML(strings.TrimSpace(huHTML), strings.TrimSpace(enHTML)))
	if err != nil {
		return "", fmt.Errorf("failed to render bilingual html: %w", err)
	}
	return strings.TrimSpace(rendered), nil
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
	htmlBody, err := joinBilingualHTML(ctx, huHTMLBody, enHTMLBody)
	if err != nil {
		return "", "", "", err
	}
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
	htmlBody, err := joinBilingualHTML(ctx, huHTMLBody, enHTMLBody)
	if err != nil {
		return "", "", "", err
	}
	return htmlBody, textBody, subject, nil
}

func (s *Service) SendInviteAccepted(ctx context.Context, to, groupName, groupID, baseURL string) error {
	htmlBody, textBody, subject, err := s.buildInviteAcceptedBodies(ctx, groupName, groupID, baseURL)
	if err != nil {
		return err
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) PreviewInviteAcceptedHTML(ctx context.Context, groupName, groupID, baseURL string) (string, error) {
	htmlBody, _, _, err := s.buildInviteAcceptedBodies(ctx, groupName, groupID, baseURL)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(htmlBody), nil
}

func (s *Service) buildInviteAcceptedBodies(ctx context.Context, groupName, groupID, baseURL string) (string, string, string, error) {
	ctx = emailContext(ctx)
	huCtx := contextWithLocale(ctx, "hu")
	enCtx := contextWithLocale(ctx, "en")

	link := groupLink(baseURL, groupID)

	subject := ctxi18ncore.T(enCtx, "email.invite_accepted.subject", groupName)
	huTextBody, err := utils.RenderHTML(huCtx, InviteAcceptedText(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite accepted text template: %w", err)
	}
	enTextBody, err := utils.RenderHTML(enCtx, InviteAcceptedText(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite accepted text template: %w", err)
	}
	huHTMLBody, err := utils.RenderHTML(huCtx, InviteAcceptedHTML(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite accepted HTML template: %w", err)
	}
	enHTMLBody, err := utils.RenderHTML(enCtx, InviteAcceptedHTML(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render invite accepted HTML template: %w", err)
	}

	textBody := joinBilingualText(huTextBody, enTextBody)
	htmlBody, err := joinBilingualHTML(ctx, huHTMLBody, enHTMLBody)
	if err != nil {
		return "", "", "", err
	}
	return htmlBody, textBody, subject, nil
}

func (s *Service) SendGroupCreated(ctx context.Context, to, groupName, groupID, baseURL string) error {
	htmlBody, textBody, subject, err := s.buildGroupCreatedBodies(ctx, groupName, groupID, baseURL)
	if err != nil {
		return err
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) PreviewGroupCreatedHTML(ctx context.Context, groupName, groupID, baseURL string) (string, error) {
	htmlBody, _, _, err := s.buildGroupCreatedBodies(ctx, groupName, groupID, baseURL)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(htmlBody), nil
}

func (s *Service) buildGroupCreatedBodies(ctx context.Context, groupName, groupID, baseURL string) (string, string, string, error) {
	ctx = emailContext(ctx)
	huCtx := contextWithLocale(ctx, "hu")
	enCtx := contextWithLocale(ctx, "en")

	link := groupLink(baseURL, groupID)

	subject := ctxi18ncore.T(enCtx, "email.group_created.subject", groupName)
	huTextBody, err := utils.RenderHTML(huCtx, GroupCreatedText(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render group created text template: %w", err)
	}
	enTextBody, err := utils.RenderHTML(enCtx, GroupCreatedText(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render group created text template: %w", err)
	}
	huHTMLBody, err := utils.RenderHTML(huCtx, GroupCreatedHTML(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render group created HTML template: %w", err)
	}
	enHTMLBody, err := utils.RenderHTML(enCtx, GroupCreatedHTML(groupName, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render group created HTML template: %w", err)
	}

	textBody := joinBilingualText(huTextBody, enTextBody)
	htmlBody, err := joinBilingualHTML(ctx, huHTMLBody, enHTMLBody)
	if err != nil {
		return "", "", "", err
	}
	return htmlBody, textBody, subject, nil
}

func (s *Service) SendRoleUpgradedToAdmin(ctx context.Context, to, groupName, groupID, baseURL string) error {
	htmlBody, textBody, subject, err := s.buildRoleChangeBodies(ctx, groupName, groupID, baseURL, "role_upgraded")
	if err != nil {
		return err
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) PreviewRoleUpgradedToAdminHTML(ctx context.Context, groupName, groupID, baseURL string) (string, error) {
	htmlBody, _, _, err := s.buildRoleChangeBodies(ctx, groupName, groupID, baseURL, "role_upgraded")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(htmlBody), nil
}

func (s *Service) SendRoleDowngradedToViewer(ctx context.Context, to, groupName, groupID, baseURL string) error {
	htmlBody, textBody, subject, err := s.buildRoleChangeBodies(ctx, groupName, groupID, baseURL, "role_downgraded")
	if err != nil {
		return err
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) PreviewRoleDowngradedToViewerHTML(ctx context.Context, groupName, groupID, baseURL string) (string, error) {
	htmlBody, _, _, err := s.buildRoleChangeBodies(ctx, groupName, groupID, baseURL, "role_downgraded")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(htmlBody), nil
}

func (s *Service) buildRoleChangeBodies(ctx context.Context, groupName, groupID, baseURL, key string) (string, string, string, error) {
	ctx = emailContext(ctx)
	huCtx := contextWithLocale(ctx, "hu")
	enCtx := contextWithLocale(ctx, "en")

	link := groupLink(baseURL, groupID)

	subject := ctxi18ncore.T(enCtx, "email."+key+".subject", groupName)

	var huTextBody, enTextBody, huHTMLBody, enHTMLBody string
	var err error

	switch key {
	case "role_upgraded":
		huTextBody, err = utils.RenderHTML(huCtx, RoleUpgradedToAdminText(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-upgraded text template: %w", err)
		}
		enTextBody, err = utils.RenderHTML(enCtx, RoleUpgradedToAdminText(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-upgraded text template: %w", err)
		}
		huHTMLBody, err = utils.RenderHTML(huCtx, RoleUpgradedToAdminHTML(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-upgraded HTML template: %w", err)
		}
		enHTMLBody, err = utils.RenderHTML(enCtx, RoleUpgradedToAdminHTML(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-upgraded HTML template: %w", err)
		}
	case "role_downgraded":
		huTextBody, err = utils.RenderHTML(huCtx, RoleDowngradedToViewerText(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-downgraded text template: %w", err)
		}
		enTextBody, err = utils.RenderHTML(enCtx, RoleDowngradedToViewerText(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-downgraded text template: %w", err)
		}
		huHTMLBody, err = utils.RenderHTML(huCtx, RoleDowngradedToViewerHTML(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-downgraded HTML template: %w", err)
		}
		enHTMLBody, err = utils.RenderHTML(enCtx, RoleDowngradedToViewerHTML(groupName, link))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render role-downgraded HTML template: %w", err)
		}
	default:
		return "", "", "", fmt.Errorf("unsupported role change key: %s", key)
	}

	textBody := joinBilingualText(huTextBody, enTextBody)
	htmlBody, err := joinBilingualHTML(ctx, huHTMLBody, enHTMLBody)
	if err != nil {
		return "", "", "", err
	}

	return htmlBody, textBody, subject, nil
}

func (s *Service) SendAccessRemoved(ctx context.Context, to, groupName string, adminEmails []string, baseURL string) error {
	htmlBody, textBody, subject, err := s.buildAccessRemovedBodies(ctx, groupName, adminEmails, baseURL)
	if err != nil {
		return err
	}

	return s.Send(to, subject, strings.TrimSpace(textBody), strings.TrimSpace(htmlBody))
}

func (s *Service) PreviewAccessRemovedHTML(ctx context.Context, groupName string, adminEmails []string, baseURL string) (string, error) {
	htmlBody, _, _, err := s.buildAccessRemovedBodies(ctx, groupName, adminEmails, baseURL)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(htmlBody), nil
}

func normalizeEmails(emails []string) []string {
	seen := make(map[string]struct{}, len(emails))
	normalized := make([]string, 0, len(emails))
	for _, email := range emails {
		trimmed := strings.TrimSpace(email)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	sort.Strings(normalized)
	return normalized
}

func (s *Service) buildAccessRemovedBodies(ctx context.Context, groupName string, adminEmails []string, baseURL string) (string, string, string, error) {
	ctx = emailContext(ctx)
	huCtx := contextWithLocale(ctx, "hu")
	enCtx := contextWithLocale(ctx, "en")

	link := dashboardLink(baseURL)
	admins := normalizeEmails(adminEmails)

	subject := ctxi18ncore.T(enCtx, "email.access_removed.subject", groupName)
	if len(admins) == 0 {
		return "", "", "", fmt.Errorf("no admin emails available for access-removed email")
	}

	huTextBody, err := utils.RenderHTML(huCtx, AccessRemovedText(groupName, admins, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render access-removed text template: %w", err)
	}
	enTextBody, err := utils.RenderHTML(enCtx, AccessRemovedText(groupName, admins, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render access-removed text template: %w", err)
	}
	huHTMLBody, err := utils.RenderHTML(huCtx, AccessRemovedHTML(groupName, admins, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render access-removed HTML template: %w", err)
	}
	enHTMLBody, err := utils.RenderHTML(enCtx, AccessRemovedHTML(groupName, admins, link))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render access-removed HTML template: %w", err)
	}

	textBody := joinBilingualText(huTextBody, enTextBody)
	htmlBody, err := joinBilingualHTML(ctx, huHTMLBody, enHTMLBody)
	if err != nil {
		return "", "", "", err
	}

	return htmlBody, textBody, subject, nil
}
