package home

import (
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func sessionUser(c echo.Context) (bool, string) {
	if cookie, err := c.Cookie(utils.SessionCookieName); err == nil && cookie.Value != "" {
		if session, err := db.Qry.GetUserSessionByToken(c.Request().Context(), cookie.Value); err == nil {
			if user, err := db.Qry.GetUserByID(c.Request().Context(), session.UserID); err == nil {
				return true, user.Email
			}

			return true, ""
		}
	}

	return false, ""
}

// Index renders the home page with welcome message.
func (h *Home) Index(c echo.Context) error {
	utils.EnsureTabID(c)

	data := h.Data(c.Request().Context())
	data.IsAuthenticated, data.UserEmail = sessionUser(c)
	return utils.RenderPage(c, HomeIndex(data))
}

func (h *Home) Pricing(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := h.LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_pricing"), ctxi18n.T(ctx, "legal.pricing"))
	data.IsAuthenticated, data.UserEmail = sessionUser(c)
	return utils.RenderPage(c, HomePricing(data))
}

func (h *Home) TermsAndConditions(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := h.LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_terms"), ctxi18n.T(ctx, "legal.terms_and_conditions"))
	data.IsAuthenticated, data.UserEmail = sessionUser(c)
	return utils.RenderPage(c, HomeTermsAndConditions(data))
}

func (h *Home) PrivacyPolicy(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := h.LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_privacy"), ctxi18n.T(ctx, "legal.privacy_policy"))
	data.IsAuthenticated, data.UserEmail = sessionUser(c)
	return utils.RenderPage(c, HomePrivacyPolicy(data))
}

func (h *Home) RefundPolicy(c echo.Context) error {
	utils.EnsureTabID(c)

	ctx := c.Request().Context()
	data := h.LegalDataWithTitle(ctx, ctxi18n.T(ctx, "legal.page_title_refund"), ctxi18n.T(ctx, "legal.refund_policy"))
	data.IsAuthenticated, data.UserEmail = sessionUser(c)
	return utils.RenderPage(c, HomeRefundPolicy(data))
}
