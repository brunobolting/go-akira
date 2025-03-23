package theme

import (
	"akira/internal/entity"
	"context"
	"net/http"
)

var _ entity.ThemeService = (*Service)(nil)

type Service struct {
	logger entity.Logger
	ctx    context.Context
}

func NewService(ctx context.Context, logger entity.Logger) *Service {
	return &Service{ctx: ctx, logger: logger}
}

func (s *Service) SetThemeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		theme := string(entity.ThemeDim)
		cookie, err := r.Cookie(entity.COOKIE_THEME_NAME)
		if err == nil && cookie != nil && cookie.Value != "" {
			theme = cookie.Value
		}
		if !isValidTheme(theme) {
			theme = string(entity.ThemeDim)
		}
		s.SetThemeCookie(w, theme)
		ctx := context.WithValue(r.Context(), entity.COOKIE_THEME_NAME, theme)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) SetThemeCookie(w http.ResponseWriter, theme string) {
	cookie := &http.Cookie{
		Name:     entity.COOKIE_THEME_NAME,
		Value:    theme,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func isValidTheme(theme string) bool {
	supportedThemes := map[string]bool{
		string(entity.ThemeLight):   true,
		string(entity.ThemeDark):    true,
		string(entity.ThemeDim):     true,
		string(entity.ThemeCupcake): true,
	}
	return supportedThemes[theme]
}
