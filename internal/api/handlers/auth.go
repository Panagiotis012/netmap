package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookie = "nm_session"
const sessionTTL = 30 * 24 * time.Hour // 30 days

type SessionStore interface {
	Create(ctx context.Context, token string, ttl time.Duration) error
	Validate(ctx context.Context, token string) (bool, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type AuthHandler struct {
	config  ConfigRepo
	session SessionStore
}

func NewAuthHandler(config ConfigRepo, session SessionStore) *AuthHandler {
	return &AuthHandler{config: config, session: session}
}

// IsSetup returns true if a password has been configured.
func (h *AuthHandler) IsSetup(ctx context.Context) bool {
	return h.config.Get(ctx, "password_hash") != ""
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	setup := h.IsSetup(r.Context())
	authenticated := h.checkCookie(r)
	writeJSON(w, http.StatusOK, map[string]bool{
		"setup":         setup,
		"authenticated": authenticated,
	})
}

func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	if h.IsSetup(r.Context()) {
		writeError(w, http.StatusConflict, "password already set")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash failed")
		return
	}
	if err := h.config.Set(r.Context(), "password_hash", string(hash)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.issueSession(w, r)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	hash := h.config.Get(r.Context(), "password_hash")
	if hash == "" {
		writeError(w, http.StatusConflict, "not set up")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "incorrect password")
		return
	}
	h.issueSession(w, r)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err == nil {
		_ = h.session.Delete(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", MaxAge: -1,
		Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) issueSession(w http.ResponseWriter, r *http.Request) {
	token := randomToken()
	_ = h.session.DeleteExpired(r.Context())
	if err := h.session.Create(r.Context(), token, sessionTTL); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		MaxAge:   int(sessionTTL.Seconds()),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
}

func (h *AuthHandler) checkCookie(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	ok, _ := h.session.Validate(r.Context(), cookie.Value)
	return ok
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
