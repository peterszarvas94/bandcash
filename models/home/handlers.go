package home

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

// Index renders the home page with welcome message.
func (h *Home) Index(c echo.Context) error {
	utils.EnsureTabID(c)

	isAuthenticated := false

	// Check if user has a valid session.
	if cookie, err := c.Cookie(utils.SessionCookieName); err == nil && cookie.Value != "" {
		if session, err := db.Qry.GetUserSessionByToken(c.Request().Context(), cookie.Value); err == nil {
			// Valid session exists.
			_ = session.UserID
			isAuthenticated = true
		}
	}

	data := h.Data(c.Request().Context())
	data.IsAuthenticated = isAuthenticated
	return utils.RenderPage(c, HomeIndex(data))
}

func (h *Home) Pricing(c echo.Context) error {
	utils.EnsureTabID(c)

	data := h.LegalDataWithTitle(c.Request().Context(), "Bandcash Pricing", "Pricing")
	return utils.RenderPage(c, HomePricing(data))
}

func (h *Home) TermsAndConditions(c echo.Context) error {
	utils.EnsureTabID(c)

	data := h.LegalDataWithTitle(c.Request().Context(), "Terms and Conditions - Bandcash", "Terms and Conditions")
	return utils.RenderPage(c, HomeTermsAndConditions(data))
}

func (h *Home) PrivacyPolicy(c echo.Context) error {
	utils.EnsureTabID(c)

	data := h.LegalDataWithTitle(c.Request().Context(), "Privacy Policy - Bandcash", "Privacy Policy")
	return utils.RenderPage(c, HomePrivacyPolicy(data))
}

func (h *Home) RefundPolicy(c echo.Context) error {
	utils.EnsureTabID(c)

	data := h.LegalDataWithTitle(c.Request().Context(), "Refund Policy - Bandcash", "Refund Policy")
	return utils.RenderPage(c, HomeRefundPolicy(data))
}
