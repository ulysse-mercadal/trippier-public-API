// Package auth exposes HTTP handlers for authentication endpoints.
package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	mw "github.com/trippier/auth-api/internal/middleware"
)

// Handler exposes auth routes.
type Handler struct {
	svc    *Service
	appURL string
}

// NewHandler builds a Handler wired to the given service svc and the
// frontend base URL appURL, and returns the configured Handler.
func NewHandler(svc *Service, appURL string) *Handler {
	return &Handler{svc: svc, appURL: appURL}
}

// RegisterRoutes mounts all auth endpoints on r, protecting /me with
// jwtAuth and rate-limiting login and registration with loginLimiter and
// registerLimiter respectively.
func (h *Handler) RegisterRoutes(r gin.IRouter, jwtAuth gin.HandlerFunc, loginLimiter gin.HandlerFunc, registerLimiter gin.HandlerFunc) {
	r.POST("/register", registerLimiter, h.register)
	r.POST("/verify-code", registerLimiter, h.verifyCode)
	r.POST("/resend-code", registerLimiter, h.resendCode)
	r.POST("/login", loginLimiter, h.login)
	r.GET("/me", jwtAuth, h.me)
}

// register handles POST /auth/register: it creates an unverified account
// and sends a 6-digit OTP, using c for the request body and to write the
// JSON response.
func (h *Handler) register(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Register(c.Request.Context(), body.Email, body.Password, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		case errors.Is(err, ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "check your email for a 6-digit verification code"})
}

// verifyCode handles POST /auth/verify-code: it checks the OTP and returns
// a JWT on success, using c for the request body and to write the JSON
// response.
func (h *Handler) verifyCode(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.svc.VerifyCode(c.Request.Context(), body.Email, body.Code)
	if err != nil {
		if errors.Is(err, ErrBadToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// resendCode handles POST /auth/resend-code: it generates a new OTP for an
// unverified account, using c for the request body and to write the JSON
// response.
func (h *Handler) resendCode(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ResendCode(c.Request.Context(), body.Email, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		if errors.Is(err, ErrAlreadyVerified) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already verified or not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification code resent"})
}

// login handles POST /auth/login: it verifies credentials and returns a
// signed JWT on success, using c for the request body and to write the
// JSON response.
func (h *Handler) login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrBadCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case errors.Is(err, ErrNotVerified):
			c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// me handles GET /auth/me: it returns the profile of the authenticated
// user, using c to identify the caller and to write the JSON response.
func (h *Handler) me(c *gin.Context) {
	userID := c.GetString(mw.UserIDKey)
	user, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, user)
}
