package billing

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	internalbilling "bandcash/internal/billing"
	"bandcash/internal/flags"
	"bandcash/internal/utils"
)

func LemonWebhook(c echo.Context) error {
	paymentsEnabled, err := flags.IsPaymentEnabled(c.Request().Context())
	if err != nil {
		slog.Error("billing.webhook: failed to read payment flag", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if !paymentsEnabled {
		slog.Info("billing.webhook: ignored because payments are disabled")
		return c.NoContent(http.StatusOK)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Warn("billing.webhook: failed to read body", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	slog.Info("billing.webhook: received request", "content_length", len(body))

	secret := utils.Env().LemonWebhookSecret
	if !internalbilling.VerifyWebhookSignature(body, c.Request().Header.Get("X-Signature"), secret) {
		slog.Warn("billing.webhook: signature verification failed", "has_signature_header", c.Request().Header.Get("X-Signature") != "", "has_endpoint_secret", secret != "")
		return c.NoContent(http.StatusUnauthorized)
	}
	slog.Info("billing.webhook: signature verification passed")

	processed, err := internalbilling.ProcessWebhook(c.Request().Context(), body)
	if err != nil {
		slog.Error("billing.webhook: processing failed", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if !processed {
		slog.Info("billing.webhook: ignored event")
	}

	return c.NoContent(http.StatusOK)
}
