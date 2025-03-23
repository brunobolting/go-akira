package web

import (
	"akira/internal/entity"
	"akira/internal/view/component"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

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
			logger.Error(r.Context(), "failed to handle request", err, nil)
			Render(w, r, component.Error(err.Error()))
		}
	}
}

func MakeMiddleware(h Middleware, logger entity.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := h(w, r); err != nil {
				logger.Error(r.Context(), "failed to handle middleware", err, nil)
				Render(w, r, component.Error(err.Error()))
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

func NewHandler(
	r *chi.Mux,
	user entity.UserService,
	session entity.SessionService,
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
	h.r.Get("/", MakeHandler(h.handleIndexPage, h.logger))
	h.r.Get("/auth/signin", MakeHandler(h.handleSignInPage, h.logger))
}
