package middleware

import (
	"net/http"

	"github.com/CAFxX/httpcompression"
	"github.com/CAFxX/httpcompression/contrib/andybalholm/brotli"
	"github.com/labstack/echo/v4"
)

// Compression returns an Echo middleware that compresses responses using Brotli.
func Compression() echo.MiddlewareFunc {
	brEnc, err := brotli.New(brotli.Options{})
	if err != nil {
		panic("failed to create brotli encoder: " + err.Error())
	}

	compress, err := httpcompression.Adapter(
		httpcompression.Compressor(brotli.Encoding, 0, brEnc),
		httpcompression.MinSize(256),
	)
	if err != nil {
		panic("failed to create compression adapter: " + err.Error())
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Create a handler that calls the next Echo handler
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.SetRequest(r)
				c.Response().Writer = w
				if err := next(c); err != nil {
					c.Error(err)
				}
			})

			// Wrap with compression and serve
			compress(h).ServeHTTP(res.Writer, req)

			return nil
		}
	}
}
