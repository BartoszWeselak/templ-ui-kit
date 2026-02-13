package i18n

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const langKey contextKey = "lang"

// WithLanguage adds language to context
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, langKey, lang)
}

// GetLanguage gets language from context
func GetLanguage(ctx context.Context) string {
	if lang, ok := ctx.Value(langKey).(string); ok {
		return lang
	}
	return "en" // default
}

// LanguageMiddleware extracts language from request and adds to context
func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := detectLanguage(r)
		ctx := WithLanguage(r.Context(), lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// detectLanguage detects language from request
func detectLanguage(r *http.Request) string {
	// 1. Check query parameter
	if lang := r.URL.Query().Get("lang"); lang != "" {
		return lang
	}

	// 2. Check cookie
	if cookie, err := r.Cookie("lang"); err == nil {
		return cookie.Value
	}

	// 3. Check Accept-Language header
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		// Simple parsing - take first language
		parts := strings.Split(acceptLang, ",")
		if len(parts) > 0 {
			lang := strings.Split(parts[0], ";")[0]
			lang = strings.Split(lang, "-")[0]
			return strings.ToLower(strings.TrimSpace(lang))
		}
	}

	// 4. Default
	return "en"
}

// SetLanguageCookie sets language cookie
func SetLanguageCookie(w http.ResponseWriter, lang string) {
	cookie := &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60, // 1 year
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}
