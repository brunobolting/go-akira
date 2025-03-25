package session

import (
	"akira/internal/entity"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var _ entity.SessionService = (*Service)(nil)

type Service struct {
	sessions   map[string]*entity.Session
	mu         sync.RWMutex
	lifetime   time.Duration
	config     *entity.CookieConfig
	gcInterval time.Duration
	repo       entity.SessionRepository
	logger     entity.Logger
	secretKey  string
}

type Options struct {
	Ctx        context.Context
	Lifetime   time.Duration
	Cookie     *entity.CookieConfig
	GCInterval time.Duration
	SecretKey  string
}

func NewService(opts Options, repository entity.SessionRepository, logger entity.Logger) *Service {
	s := &Service{
		sessions:   make(map[string]*entity.Session),
		lifetime:   opts.Lifetime,
		config:     opts.Cookie,
		gcInterval: opts.GCInterval,
		repo:       repository,
		logger:     logger,
		secretKey:  opts.SecretKey,
	}
	s.RunGC(opts.Ctx)
	return s
}

func (s *Service) CreateSession(ctx context.Context, userID string) (*entity.Session, error) {
	ID, err := s.GenerateSessionID()
	if err != nil {
		s.logger.Error(ctx, "failed to generate session ID", err, map[string]any{"userID": userID})
		return nil, err
	}
	session := entity.NewSession(ID, userID, make(map[string]any), s.lifetime)
	s.mu.Lock()
	s.sessions[ID] = session
	s.mu.Unlock()
	if err := s.repo.CreateSession(session); err != nil {
		s.logger.Error(ctx, "failed to create session", err, map[string]any{"session": session})
		return nil, err
	}
	return session, nil
}

func (s *Service) FindSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	ID, valid := s.VerifySessionID(sessionID)
	if !valid {
		s.logger.Error(ctx, "invalid session signature ID", entity.ErrInvalidSession, map[string]any{"sessionID": sessionID})
		return nil, entity.ErrInvalidSession
	}
	s.mu.RLock()
	session, exists := s.sessions[ID]
	s.mu.RUnlock()
	if !exists {
		var err error
		session, err = s.repo.FindSession(ID)
		if err != nil {
			if err == entity.ErrSessionNotFound {
				return nil, err
			}
			s.logger.Error(ctx, "failed to find session", err, map[string]any{"sessionID": ID})
			return nil, err
		}
		s.mu.Lock()
		s.sessions[ID] = session
		s.mu.Unlock()
	}
	if session == nil {
		return nil, entity.ErrSessionNotFound
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		s.DeleteSession(ctx, session.ID)
		return nil, entity.ErrSessionExpired
	}
	return session, nil
}

func (s *Service) GetSession(ctx context.Context) (*entity.Session, error) {
	session, ok := ctx.Value(entity.SESSION_NAME).(*entity.Session)
	if !ok || session == nil {
		return nil, entity.ErrSessionNotFound
	}
	return session, nil
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
	if err := s.repo.DeleteSession(sessionID); err != nil {
		s.logger.Error(ctx, "failed to delete session", err, map[string]any{"sessionID": sessionID})
		return err
	}
	return nil
}

func (s *Service) SetCookie(ctx context.Context, w http.ResponseWriter, sessionID string) {
	signedID := s.SignSessionID(sessionID)
	http.SetCookie(w, &http.Cookie{
		Name:     s.config.Name,
		Value:    signedID,
		Path:     s.config.Path,
		Domain:   s.config.Domain,
		MaxAge:   s.config.MaxAge,
		Secure:   s.config.Secure,
		HttpOnly: s.config.HttpOnly,
		SameSite: s.config.SameSite,
	})
}

func (s *Service) ClearCookie(ctx context.Context, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.config.Name,
		Value:    "",
		Path:     s.config.Path,
		Domain:   s.config.Domain,
		MaxAge:   -1,
		Secure:   s.config.Secure,
		HttpOnly: s.config.HttpOnly,
		SameSite: s.config.SameSite,
	})
}

func (s *Service) GC(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiredSessions, err := s.FindExpiredSessions(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to find expired sessions on GC", err, nil)
		return
	}
	for _, session := range expiredSessions {
		delete(s.sessions, session.ID)
		err := s.repo.DeleteSession(session.ID)
		if err != nil {
			s.logger.Error(ctx, "failed to delete expired session on GC", err, map[string]any{"session": session})
		}
	}
}

func (s *Service) RunGC(ctx context.Context) {
	go func() {
		i := s.gcInterval
		if i == 0 {
			i = 30 * time.Minute
		}
		ticker := time.NewTicker(i)
		defer ticker.Stop()
		for range ticker.C {
			s.GC(ctx)
		}
	}()
}

func (s *Service) FindExpiredSessions(ctx context.Context) ([]entity.Session, error) {
	return s.repo.GetExpiredSessions()
}

func (s *Service) GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *Service) SignSessionID(sessionID string) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(sessionID))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s.%s", sessionID, signature)
}

func (s *Service) VerifySessionID(signedID string) (string, bool) {
	parts := strings.Split(signedID, ".")
	if len(parts) != 2 {
		return "", false
	}
	sessionID := parts[0]
	expectedSignature := s.SignSessionID(sessionID)
	return sessionID, hmac.Equal([]byte(expectedSignature), []byte(signedID))
}

func (s *Service) SetSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.config.Name)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		session, err := s.FindSession(r.Context(), cookie.Value)
		if err != nil {
			s.ClearCookie(r.Context(), w)
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), entity.SESSION_NAME, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) AuthenticationRequiredMiddleware(w http.ResponseWriter, r *http.Request) error {
	_, err := s.GetSession(r.Context())
	if err != nil {
		return entity.ErrUserUnauthorized
	}
	return nil
}
