package web

import (
	"akira/internal/entity"
	"akira/internal/view/component/form"
	"errors"
	"net/http"
)

func (h *Handler) handleSignUpRequest(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	req := entity.SignUpRequest{
		Name:     r.Form.Get("name"),
		Email:    r.Form.Get("email"),
		Password: r.Form.Get("password"),
		Captcha:  r.Form.Get("cf-turnstile-response"),
	}
	user, err := h.auth.SignUp(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrUserAlreadyExists) {
			return WebError{code: http.StatusConflict, msg: "error.user.already-exists"}
		}
		if _, ok := err.(entity.RequestError); ok {
			err := err.(entity.RequestError)
			return Render(w, r, form.SignUp(form.SignUpProps{
				Name:     req.Name,
				Email:    req.Email,
				Password: req.Password,
			}, &err))
		}
		return err
	}
	s, err := h.session.CreateSession(r.Context(), user.ID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to create session", err, nil)
		return err
	}
	h.session.SetCookie(r.Context(), w, s.ID)
	return HxRedirect(w, r, "/")
}
