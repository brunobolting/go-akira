package i18n

import (
	"akira/internal/entity"
	"context"
	"net/http"
	"strings"

	"github.com/invopop/ctxi18n"
)

var _ entity.I18nService = (*Service)(nil)

type Service struct {
	ctx    context.Context
	logger entity.Logger
}

func NewService(ctx context.Context, logger entity.Logger) *Service {
	return &Service{
		ctx:    ctx,
		logger: logger,
	}
}

func (s *Service) SetLocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := string(entity.LocaleEN)
		cookie, err := r.Cookie(entity.COOKIE_LOCALE_NAME)
		if err == nil && cookie != nil && cookie.Value != "" {
			lang = cookie.Value
		} else if err == http.ErrNoCookie {
			acceptLang := r.Header.Get("Accept-Language")
			if acceptLang != "" {
				preferredLang := parseAcceptLanguage(acceptLang)
				if isValidLocale(preferredLang) {
					lang = preferredLang
				}
			}
		} else if err != nil && err != http.ErrNoCookie {
			s.logger.Error(r.Context(), "error reading locale cookie", err, map[string]any{
				"error_type": "cookie_read",
			})
		}
		s.SetLocaleCookie(w, lang)
		ctx, err := ctxi18n.WithLocale(r.Context(), lang)
		if err != nil {
			s.logger.Error(r.Context(), "failed to set locale", err, map[string]any{
				"locale": lang,
			})
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) SetLocaleCookie(w http.ResponseWriter, locale string) {
	cookie := &http.Cookie{
		Name:     entity.COOKIE_LOCALE_NAME,
		Value:    locale,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func parseAcceptLanguage(acceptLanguage string) string {
	languages := strings.Split(acceptLanguage, ",")
	if len(languages) == 0 {
		return string(entity.LocaleEN)
	}
	highest := languages[0]
	langParts := strings.Split(highest, ";")
	langWithRegion := langParts[0]
	langParts = strings.Split(langWithRegion, "-")
	return langParts[0]
}

func isValidLocale(locale string) bool {
	supportedLocales := map[string]bool{
		string(entity.LocaleEN): true,
		string(entity.LocaleBR): true,
	}
	return supportedLocales[locale]
}
