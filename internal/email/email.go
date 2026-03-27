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

type builtEmail struct {
	Subject  string
	TextBody string
	HTMLBody string
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
	return fmt.Sprintf("%s/groups", baseURL)
}

func joinBilingualText(huText, enText string) string {
	return strings.TrimSpace("English follows below.\n\n" + strings.TrimSpace(huText) + "\n\n---\n\n" + strings.TrimSpace(enText))
}

func joinBilingualSubject(huSubject, enSubject string) string {
	hu := strings.TrimSpace(huSubject)
	en := strings.TrimSpace(enSubject)
	if hu == "" {
		return en
	}
	if en == "" || hu == en {
		return hu
	}
	return hu + " / " + en
}

func subjectForLocale(ctx context.Context, locale, key string, args ...any) string {
	return ctxi18ncore.T(contextWithLocale(emailContext(ctx), locale), key, args...)
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

func trimBuiltEmail(data builtEmail) builtEmail {
	data.Subject = strings.TrimSpace(data.Subject)
	data.TextBody = strings.TrimSpace(data.TextBody)
	data.HTMLBody = strings.TrimSpace(data.HTMLBody)
	return data
}

func (s *Service) sendBuilt(ctx context.Context, to string, build func(context.Context) (builtEmail, error)) error {
	data, err := build(ctx)
	if err != nil {
		return err
	}
	data = trimBuiltEmail(data)
	return s.Send(to, data.Subject, data.TextBody, data.HTMLBody)
}

func (s *Service) previewBuilt(ctx context.Context, build func(context.Context) (builtEmail, error)) (string, string, error) {
	data, err := build(ctx)
	if err != nil {
		return "", "", err
	}
	data = trimBuiltEmail(data)
	return data.Subject, data.HTMLBody, nil
}

func buildBilingualEmail(
	ctx context.Context,
	subjectHU, subjectEN string,
	renderHUText, renderENText, renderHUHTML, renderENHTML func(context.Context) (string, error),
) (builtEmail, error) {
	ctx = emailContext(ctx)
	huCtx := contextWithLocale(ctx, "hu")
	enCtx := contextWithLocale(ctx, "en")

	huTextBody, err := renderHUText(huCtx)
	if err != nil {
		return builtEmail{}, err
	}
	enTextBody, err := renderENText(enCtx)
	if err != nil {
		return builtEmail{}, err
	}
	huHTMLBody, err := renderHUHTML(huCtx)
	if err != nil {
		return builtEmail{}, err
	}
	enHTMLBody, err := renderENHTML(enCtx)
	if err != nil {
		return builtEmail{}, err
	}

	htmlBody, err := joinBilingualHTML(ctx, huHTMLBody, enHTMLBody)
	if err != nil {
		return builtEmail{}, err
	}

	return builtEmail{
		Subject:  joinBilingualSubject(subjectHU, subjectEN),
		TextBody: joinBilingualText(huTextBody, enTextBody),
		HTMLBody: htmlBody,
	}, nil
}

func (s *Service) SendMagicLink(ctx context.Context, to, token, baseURL string) error {
	return s.sendBuilt(ctx, to, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildMagicLinkBodies(buildCtx, token, baseURL)
	})
}

func (s *Service) PreviewMagicLinkHTML(ctx context.Context, token, baseURL string) (string, string, error) {
	return s.previewBuilt(ctx, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildMagicLinkBodies(buildCtx, token, baseURL)
	})
}

func (s *Service) buildMagicLinkBodies(ctx context.Context, token, baseURL string) (builtEmail, error) {
	huLink := verifyLink(baseURL, token, "hu")
	enLink := verifyLink(baseURL, token, "en")
	return buildBilingualEmail(
		ctx,
		subjectForLocale(ctx, "hu", "email.magic_link.subject"),
		subjectForLocale(ctx, "en", "email.magic_link.subject"),
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, MagicLinkText(huLink))
			if err != nil {
				return "", fmt.Errorf("failed to render magic-link text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, MagicLinkText(enLink))
			if err != nil {
				return "", fmt.Errorf("failed to render magic-link text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, MagicLinkHTML(huLink))
			if err != nil {
				return "", fmt.Errorf("failed to render magic-link HTML template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, MagicLinkHTML(enLink))
			if err != nil {
				return "", fmt.Errorf("failed to render magic-link HTML template: %w", err)
			}
			return body, nil
		},
	)
}

func (s *Service) SendGroupInvitation(ctx context.Context, to, groupName, token, baseURL string) error {
	return s.sendBuilt(ctx, to, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildGroupInvitationBodies(buildCtx, groupName, token, baseURL)
	})
}

func (s *Service) PreviewGroupInvitationHTML(ctx context.Context, groupName, token, baseURL string) (string, string, error) {
	return s.previewBuilt(ctx, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildGroupInvitationBodies(buildCtx, groupName, token, baseURL)
	})
}

func (s *Service) buildGroupInvitationBodies(ctx context.Context, groupName, token, baseURL string) (builtEmail, error) {

	huLink := verifyLink(baseURL, token, "hu")
	enLink := verifyLink(baseURL, token, "en")

	return buildBilingualEmail(
		ctx,
		subjectForLocale(ctx, "hu", "email.invite.subject", groupName),
		subjectForLocale(ctx, "en", "email.invite.subject", groupName),
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupInvitationText(groupName, huLink))
			if err != nil {
				return "", fmt.Errorf("failed to render invite text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupInvitationText(groupName, enLink))
			if err != nil {
				return "", fmt.Errorf("failed to render invite text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupInvitationHTML(groupName, huLink))
			if err != nil {
				return "", fmt.Errorf("failed to render invite HTML template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupInvitationHTML(groupName, enLink))
			if err != nil {
				return "", fmt.Errorf("failed to render invite HTML template: %w", err)
			}
			return body, nil
		},
	)
}

func (s *Service) SendInviteAccepted(ctx context.Context, to, groupName, groupID, baseURL string) error {
	return s.sendBuilt(ctx, to, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildInviteAcceptedBodies(buildCtx, groupName, groupID, baseURL)
	})
}

func (s *Service) PreviewInviteAcceptedHTML(ctx context.Context, groupName, groupID, baseURL string) (string, string, error) {
	return s.previewBuilt(ctx, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildInviteAcceptedBodies(buildCtx, groupName, groupID, baseURL)
	})
}

func (s *Service) buildInviteAcceptedBodies(ctx context.Context, groupName, groupID, baseURL string) (builtEmail, error) {

	link := groupLink(baseURL, groupID)

	return buildBilingualEmail(
		ctx,
		subjectForLocale(ctx, "hu", "email.invite_accepted.subject", groupName),
		subjectForLocale(ctx, "en", "email.invite_accepted.subject", groupName),
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, InviteAcceptedText(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render invite accepted text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, InviteAcceptedText(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render invite accepted text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, InviteAcceptedHTML(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render invite accepted HTML template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, InviteAcceptedHTML(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render invite accepted HTML template: %w", err)
			}
			return body, nil
		},
	)
}

func (s *Service) SendGroupCreated(ctx context.Context, to, groupName, groupID, baseURL string) error {
	return s.sendBuilt(ctx, to, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildGroupCreatedBodies(buildCtx, groupName, groupID, baseURL)
	})
}

func (s *Service) PreviewGroupCreatedHTML(ctx context.Context, groupName, groupID, baseURL string) (string, string, error) {
	return s.previewBuilt(ctx, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildGroupCreatedBodies(buildCtx, groupName, groupID, baseURL)
	})
}

func (s *Service) buildGroupCreatedBodies(ctx context.Context, groupName, groupID, baseURL string) (builtEmail, error) {

	link := groupLink(baseURL, groupID)

	return buildBilingualEmail(
		ctx,
		subjectForLocale(ctx, "hu", "email.group_created.subject", groupName),
		subjectForLocale(ctx, "en", "email.group_created.subject", groupName),
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupCreatedText(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render group created text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupCreatedText(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render group created text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupCreatedHTML(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render group created HTML template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, GroupCreatedHTML(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render group created HTML template: %w", err)
			}
			return body, nil
		},
	)
}

func (s *Service) SendRoleUpgradedToAdmin(ctx context.Context, to, groupName, groupID, baseURL string) error {
	return s.sendBuilt(ctx, to, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildRoleChangeBodies(buildCtx, groupName, groupID, baseURL, "role_upgraded")
	})
}

func (s *Service) PreviewRoleUpgradedToAdminHTML(ctx context.Context, groupName, groupID, baseURL string) (string, string, error) {
	return s.previewBuilt(ctx, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildRoleChangeBodies(buildCtx, groupName, groupID, baseURL, "role_upgraded")
	})
}

func (s *Service) SendRoleDowngradedToViewer(ctx context.Context, to, groupName, groupID, baseURL string) error {
	return s.sendBuilt(ctx, to, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildRoleChangeBodies(buildCtx, groupName, groupID, baseURL, "role_downgraded")
	})
}

func (s *Service) PreviewRoleDowngradedToViewerHTML(ctx context.Context, groupName, groupID, baseURL string) (string, string, error) {
	return s.previewBuilt(ctx, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildRoleChangeBodies(buildCtx, groupName, groupID, baseURL, "role_downgraded")
	})
}

func (s *Service) buildRoleChangeBodies(ctx context.Context, groupName, groupID, baseURL, key string) (builtEmail, error) {

	link := groupLink(baseURL, groupID)
	var renderHUText, renderENText, renderHUHTML, renderENHTML func(context.Context) (string, error)

	switch key {
	case "role_upgraded":
		renderHUText = func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, RoleUpgradedToAdminText(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render role-upgraded text template: %w", err)
			}
			return body, nil
		}
		renderENText = renderHUText
		renderHUHTML = func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, RoleUpgradedToAdminHTML(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render role-upgraded HTML template: %w", err)
			}
			return body, nil
		}
		renderENHTML = renderHUHTML
	case "role_downgraded":
		renderHUText = func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, RoleDowngradedToViewerText(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render role-downgraded text template: %w", err)
			}
			return body, nil
		}
		renderENText = renderHUText
		renderHUHTML = func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, RoleDowngradedToViewerHTML(groupName, link))
			if err != nil {
				return "", fmt.Errorf("failed to render role-downgraded HTML template: %w", err)
			}
			return body, nil
		}
		renderENHTML = renderHUHTML
	default:
		return builtEmail{}, fmt.Errorf("unsupported role change key: %s", key)
	}

	return buildBilingualEmail(
		ctx,
		subjectForLocale(ctx, "hu", "email."+key+".subject", groupName),
		subjectForLocale(ctx, "en", "email."+key+".subject", groupName),
		renderHUText,
		renderENText,
		renderHUHTML,
		renderENHTML,
	)
}

func (s *Service) SendAccessRemoved(ctx context.Context, to, groupName string, adminEmails []string, baseURL string) error {
	return s.sendBuilt(ctx, to, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildAccessRemovedBodies(buildCtx, groupName, adminEmails, baseURL)
	})
}

func (s *Service) PreviewAccessRemovedHTML(ctx context.Context, groupName string, adminEmails []string, baseURL string) (string, string, error) {
	return s.previewBuilt(ctx, func(buildCtx context.Context) (builtEmail, error) {
		return s.buildAccessRemovedBodies(buildCtx, groupName, adminEmails, baseURL)
	})
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

func (s *Service) buildAccessRemovedBodies(ctx context.Context, groupName string, adminEmails []string, baseURL string) (builtEmail, error) {

	link := dashboardLink(baseURL)
	admins := normalizeEmails(adminEmails)
	if len(admins) == 0 {
		return builtEmail{}, fmt.Errorf("no admin emails available for access-removed email")
	}

	return buildBilingualEmail(
		ctx,
		subjectForLocale(ctx, "hu", "email.access_removed.subject", groupName),
		subjectForLocale(ctx, "en", "email.access_removed.subject", groupName),
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, AccessRemovedText(groupName, admins, link))
			if err != nil {
				return "", fmt.Errorf("failed to render access-removed text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, AccessRemovedText(groupName, admins, link))
			if err != nil {
				return "", fmt.Errorf("failed to render access-removed text template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, AccessRemovedHTML(groupName, admins, link))
			if err != nil {
				return "", fmt.Errorf("failed to render access-removed HTML template: %w", err)
			}
			return body, nil
		},
		func(localCtx context.Context) (string, error) {
			body, err := utils.RenderHTML(localCtx, AccessRemovedHTML(groupName, admins, link))
			if err != nil {
				return "", fmt.Errorf("failed to render access-removed HTML template: %w", err)
			}
			return body, nil
		},
	)
}
