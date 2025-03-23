package entity

import (
	"net/http"
)

const COOKIE_THEME_NAME = "akira_preferred_theme"

type Theme struct {
	Name  string
	Value string
}

const ThemeDefault = "dim"

type ThemeService interface {
	SetThemeMiddleware(next http.Handler) http.Handler
	SetThemeCookie(w http.ResponseWriter, theme string)
}

func IsValidTheme(theme string) bool {
	for _, t := range GetThemes() {
		if t.Value == theme {
			return true
		}
	}
	return false
}

func GetThemes() []Theme {
	return []Theme {
		{
			Name:  "Light",
			Value: "light",
		},
		{
			Name:  "Dark",
			Value: "dark",
		},
		{
			Name:  "Dim",
			Value: "dim",
		},
		{
			Name:  "Cupcake",
			Value: "cupcake",
		},
	}
}
