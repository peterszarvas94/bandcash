package billing

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	internalbilling "bandcash/internal/billing"
	"bandcash/internal/utils"
)

func LemonWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Warn("billing.webhook: failed to read body", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}

	secret := utils.Env().LemonWebhookSecret
	if !internalbilling.VerifyWebhookSignature(body, c.Request().Header.Get("X-Signature"), secret) {
		slog.Warn("billing.webhook: signature verification failed")
		return c.NoContent(http.StatusUnauthorized)
	}

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
