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
