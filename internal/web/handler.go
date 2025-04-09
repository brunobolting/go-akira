package web

import (
	"akira/internal/entity"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/invopop/ctxi18n/i18n"
)

type WebError struct {
	code int
	msg  string
}

func (e WebError) Error() string {
	return string(e.msg)
}

type WebHandler func(w http.ResponseWriter, r *http.Request) error

type Middleware func(w http.ResponseWriter, r *http.Request) error

type Options struct {
	AllowedOrigins []string
}

type Handler struct {
	r       *chi.Mux
	mu      *sync.Mutex
	user    entity.UserService
	session entity.SessionService
	auth    entity.AuthService
	logger  entity.Logger
	i18n    entity.I18nService
	theme   entity.ThemeService
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func MakeHandler(h WebHandler, logger entity.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			if _, ok := err.(WebError); ok {
				http.Error(w, i18n.T(r.Context(), err.Error()), err.(WebError).code)
				return
			}
			logger.Error(r.Context(), "failed to handle request", err, nil)
			http.Error(w, i18n.T(r.Context(), "error.unexpected-error"), http.StatusInternalServerError)
		}
	}
}

func MakeMiddleware(h Middleware, logger entity.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := h(w, r); err != nil {
				if err == entity.ErrUserUnauthorized {
					HxRedirect(w, r, fmt.Sprintf("/auth/signin?error=%s", i18n.T(r.Context(), err.Error())))
					return
				}
				if _, ok := err.(WebError); ok {
					http.Error(w, i18n.T(r.Context(), err.Error()), err.(WebError).code)
					return
				}
				logger.Error(r.Context(), "failed to handle middleware", err, nil)
				http.Error(w, i18n.T(r.Context(), "error.unexpected-error"), http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Render(w http.ResponseWriter, r *http.Request, c templ.Component) error {
	return c.Render(r.Context(), w)
}

func HxRedirect(w http.ResponseWriter, r *http.Request, url string) error {
	if len(r.Header.Get("HX-Request")) > 0 {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
	return nil
}

func HxTrigger(w http.ResponseWriter, event string, data map[string]string) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	w.Header().Set("HX-Trigger", event)
	w.Header().Set("HX-Trigger-Content-Type", "application/json")
	w.Write(payload)
	return nil
}

func HxRefresh(w http.ResponseWriter) error {
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
	return nil
}

func NewHandler(
	r *chi.Mux,
	user entity.UserService,
	session entity.SessionService,
	auth entity.AuthService,
	logger entity.Logger,
	i18n entity.I18nService,
	theme entity.ThemeService,
	opts Options,
) *Handler {
	h := &Handler{
		r:       r,
		mu:      &sync.Mutex{},
		user:    user,
		session: session,
		auth:    auth,
		logger:  logger,
		i18n:    i18n,
		theme:   theme,
	}
	h.r.Use(chi_middleware.Logger)
	h.r.Use(chi_middleware.RequestID, chi_middleware.Recoverer)
	h.r.Use(h.session.SetSessionMiddleware)
	h.r.Use(h.i18n.SetLocaleMiddleware)
	h.r.Use(h.theme.SetThemeMiddleware)
	h.r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   opts.AllowedOrigins,
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	h.MakeRoutes()
	return h
}

func (h *Handler) MakeRoutes() {
	h.r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP)
	h.r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	h.r.Route("/", func(r chi.Router) {
		r.Use(MakeMiddleware(h.session.AuthenticationRequiredMiddleware, h.logger))
		r.Get("/", MakeHandler(h.handleIndexPage, h.logger))
		r.Get("/collection/create", MakeHandler(h.handleCreateCollectionPage, h.logger))
	})
	h.r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, i18n.T(r.Context(), "error.unexpected-error"), http.StatusInternalServerError)
	})
	h.r.Route("/auth", func(r chi.Router) {
		r.Get("/signup", MakeHandler(h.handleSignUpPage, h.logger))
		r.Post("/signup", MakeHandler(h.handleSignUpRequest, h.logger))
		r.Get("/signin", MakeHandler(h.handleSignInPage, h.logger))
		r.Post("/signin", MakeHandler(h.handleSignInRequest, h.logger))
		r.Get("/signout", MakeHandler(h.handleSignOutRequest, h.logger))
	})
	h.r.Route("/api", func(r chi.Router) {
		r.Post("/change-theme", MakeHandler(h.handleChangeTheme, h.logger))
		r.Post("/change-locale", MakeHandler(h.handleChangeLocale, h.logger))
	})
}
