package web

import (
	"akira/internal/view/page"
	"net/http"
)

func (h *Handler) handleIndexPage(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, page.Index())
}

func (h *Handler) handleSignUpPage(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, page.SignUp())
}
