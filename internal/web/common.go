package web

import (
	"akira/internal/entity"
	"net/http"
)

func (h *Handler) handleChangeTheme(w http.ResponseWriter, r *http.Request) error {
	theme := r.FormValue("theme")
	if !entity.IsValidTheme(theme) {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	h.theme.SetThemeCookie(w, theme)
	return HxTrigger(w, "themeChanged", map[string]string{"theme": theme})
}

func (h *Handler) handleChangeLocale(w http.ResponseWriter, r *http.Request) error {
	locale := r.FormValue("locale")
	if !entity.IsValidLocale(locale) {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	h.i18n.SetLocaleCookie(w, locale)
	return HxRefresh(w)
}
