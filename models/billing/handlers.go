package billing

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	internalbilling "bandcash/internal/billing"
	"bandcash/internal/utils"
)

const webhookSignatureTolerance = 5 * time.Minute

func PaddleWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Warn("billing.webhook: failed to read body", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}

	secret := utils.Env().PaddleWebhookSecret
	if !internalbilling.VerifyWebhookSignature(body, c.Request().Header.Get("Paddle-Signature"), secret, webhookSignatureTolerance) {
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
