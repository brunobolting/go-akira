package web

import (
	"akira/internal/view/component/form"
	"akira/internal/view/page"
	"net/http"
)

func (h *Handler) handleIndexPage(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, page.Index())
}

func (h *Handler) handleSignUpPage(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return Render(w, r, page.SignUp(form.SignUpProps{
		Name:     "",
		Email:    "",
		Password: "",
	}, nil))
}

func (h *Handler) handleSignInPage(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	return Render(w, r, page.SignIn(page.SignInProps{
		Email:    email,
		Password: password,
	}, nil))
}
