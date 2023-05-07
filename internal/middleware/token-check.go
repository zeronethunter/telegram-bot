package middlewareBot

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func TokenCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestToken := c.Request().Header.Get("X-Telegram-Bot-Api-Secret-Token")
			if requestToken != os.Getenv("WEBHOOK_SECRET_TOKEN") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}
	}
}
