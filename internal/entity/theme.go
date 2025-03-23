package entity

import "net/http"

const COOKIE_THEME_NAME = "akira_preferred_theme"

type Theme string

const (
	ThemeLight   Theme = "light"
	ThemeDark    Theme = "dark"
	ThemeDim     Theme = "dim"
	ThemeCupcake Theme = "cupcake"
)

type ThemeService interface {
	SetThemeMiddleware(next http.Handler) http.Handler
	SetThemeCookie(w http.ResponseWriter, theme string)
}
