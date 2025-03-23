package entity

import "net/http"

const COOKIE_LOCALE_NAME = "akira_preferred_language"

type Locale string

const (
	LocaleEN Locale = "en"
	LocaleBR Locale = "pt-BR"
)

type I18nService interface {
	SetLocaleMiddleware(next http.Handler) http.Handler
	SetLocaleCookie(w http.ResponseWriter, locale string)
}

func IsValidLocale(lang string) bool {
	for _, l := range GetLanguages() {
		if string(l) == lang {
			return true
		}
	}
	return false
}

func GetLanguages() []Locale {
	return []Locale{
		LocaleEN,
		LocaleBR,
	}
}
