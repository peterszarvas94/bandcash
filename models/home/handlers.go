package home

import (
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/store"
)

func sessionUser(c echo.Context) (bool, string) {
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err != nil || cookie.Value == "" {
		return false, ""
	}

	session, err := authstore.GetUserSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return false, ""
	}

	user, err := authstore.GetUserByID(c.Request().Context(), session.UserID)
	if err != nil {
		return false, ""
	}

	syncPreferredLangFromQuery(c, user.ID, user.PreferredLang)
	return true, user.Email
}

func syncPreferredLangFromQuery(c echo.Context, userID string, currentPreferredLang string) {
	rawLang := strings.TrimSpace(c.QueryParam("lang"))
	if rawLang == "" {
		return
	}

	lang := appi18n.NormalizeLocale(rawLang)
	if appi18n.NormalizeLocale(currentPreferredLang) == lang {
		return
	}

	_ = authstore.UpdateUserPreferredLang(c.Request().Context(), authstore.UpdateUserPreferredLangParams{
		ID:            userID,
		PreferredLang: lang,
	})
}

// Index renders the home page with welcome message.
func Index(c echo.Context) error {
	utils.EnsureTabID(c)

	data := Data(c.Request().Context())
	data.IsAuthenticated, _ = sessionUser(c)
	data.IsSuperAdmin = false
	return utils.RenderPage(c, HomeIndex(data))
}

func Pricing(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_pricing"), ctxi18n.T(ctx, "legal.pricing"))
	data.IsAuthenticated, _ = sessionUser(c)
	data.IsSuperAdmin = false
	return utils.RenderPage(c, HomePricing(data))
}

func TermsAndConditions(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_terms"), ctxi18n.T(ctx, "legal.terms_and_conditions"))
	data.IsAuthenticated, _ = sessionUser(c)
	data.IsSuperAdmin = false
	return utils.RenderPage(c, HomeTermsAndConditions(data))
}

func PrivacyPolicy(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_privacy"), ctxi18n.T(ctx, "legal.privacy_policy"))
	data.IsAuthenticated, _ = sessionUser(c)
	data.IsSuperAdmin = false
	return utils.RenderPage(c, HomePrivacyPolicy(data))
}

func RefundPolicy(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_refund"), ctxi18n.T(ctx, "legal.refund_policy"))
	data.IsAuthenticated, _ = sessionUser(c)
	data.IsSuperAdmin = false
	return utils.RenderPage(c, HomeRefundPolicy(data))
}
