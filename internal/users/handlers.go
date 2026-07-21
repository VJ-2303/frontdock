package users

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/VJ-2303/frontdock/internal/auth"
	"github.com/VJ-2303/frontdock/internal/config"
	"github.com/VJ-2303/frontdock/internal/httpx"
	"github.com/VJ-2303/frontdock/internal/queue"
)

type Handler struct {
	users *Store
	cfg   *config.Config
	pub   *queue.Publisher
}

func NewHandler(user *Store, cfg *config.Config, pub *queue.Publisher) *Handler {
	return &Handler{
		users: user,
		cfg:   cfg,
		pub:   pub,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !strings.Contains(req.Email, "@") || len(req.Email) > 254 {
		httpx.Error(w, 422, "validation_failed", "invalid email address")
		return
	}
	if len(req.Password) < 8 || len(req.Password) > 128 {
		httpx.Error(w, 422, "validation_failed", "password must be 8-128 characters")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "could not hash password")
		return
	}
	u, err := h.users.Create(r.Context(), req.Email, hash)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			httpx.Error(w, 409, "email_taken", "that email is already registered")
		} else {
			httpx.Error(w, 500, "internal_error", "could not create user")
		}
		return
	}
	raw, hash, err := auth.NewToken()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "could not create verification token")
		return
	}

	if err := h.users.CreateVerificationToken(r.Context(), u.ID, hash, time.Hour*24); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "could not create verification token")
		return
	}

	if err := h.pub.Publish(r.Context(), queue.RoutingEmailSend, queue.EmailMessage{
		Type:     "email.send",
		To:       u.Email,
		Template: "verify",
		Data: map[string]any{
			"PublicAPIURL": h.cfg.PublicAPIURL,
			"Token":        raw,
		},
	}); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "could not create verification token")
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"user_id":    u.ID,
		"user_email": u.Email,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	u, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		httpx.Error(w, http.StatusUnauthorized, "invalid_credentials", "incorrect email or password")
		return
	}
	token, err := auth.IssueToken(h.cfg.JWTSecret, u.ID, u.isVerified(), h.cfg.JWTTTL)
	if err != nil {
		httpx.Error(w, 500, "internal_error", "could not issue token")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": u.ID, "email": u.Email},
	})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		httpx.Error(w, http.StatusBadRequest, "verification_token_missing", "Verification token is missing in the email")
		return
	}
	hash := auth.HashToken(token)

	id, err := h.users.ConsumeVerificationToken(r.Context(), hash)
	if err != nil {
		if errors.Is(err, ErrVerificationTokenInvalid) {
			httpx.Error(w, http.StatusUnauthorized, "invalid_verification_token", "verification is invalid it may be wrong or expired")
		} else {
			httpx.Error(w, http.StatusInternalServerError, "internal_error", "please try again later")
		}
		return
	}
	jwtToken, err := auth.IssueToken(h.cfg.JWTSecret, id, true, h.cfg.JWTTTL)
	if err != nil {
		httpx.Error(w, 500, "internal_error", "could not issue token")
		return
	}
	httpx.JSON(
		w, http.StatusOK,
		map[string]any{
			"token":   jwtToken,
			"message": "Email Verified Successfully, you can login now",
		},
	)
}
