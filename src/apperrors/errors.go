package apperrors

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"templ-ui-kit/src/i18n"
)

func Handler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	messageKey := "error.internal"

	var he *echo.HTTPError
	if stderrors.As(err, &he) {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			messageKey = msg
		}
	}

	c.Logger().Errorf("HTTP %d: %v", code, err)

	if c.Request().Method == http.MethodHead {
		_ = c.NoContent(code)
		return
	}

	ctx := c.Request().Context()
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(code)
	if rErr := ErrorPage(code, messageKey).Render(ctx, c.Response().Writer); rErr != nil {
		c.Logger().Errorf("failed to render error page: %v", rErr)
	}
}

func RenderTempl(c echo.Context, status int, t templ.Component) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(status)
	if err := t.Render(c.Request().Context(), c.Response().Writer); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error.render")
	}
	return nil
}

func Tr(lang, key string) string {
	return i18n.GetTranslator().T(lang, key)
}

func StatusTitle(lang string, code int) string {
	key := fmt.Sprintf("error.%d.title", code)
	if title := Tr(lang, key); title != key {
		return title
	}
	return http.StatusText(code)
}

func StatusMessage(lang, messageKey string) string {
	if msg := Tr(lang, messageKey); msg != messageKey {
		return msg
	}
	return Tr(lang, "error.generic")
}

func StatusEmoji(code int) string {
	switch {
	case code == http.StatusNotFound:
		return "🔍"
	case code == http.StatusForbidden:
		return "🔒"
	case code == http.StatusUnauthorized:
		return "🔑"
	case code == http.StatusTooManyRequests:
		return "🚦"
	case code >= 500:
		return "💥"
	default:
		return "⚠️"
	}
}
