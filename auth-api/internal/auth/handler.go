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

// NewHandler creates a Handler.
func NewHandler(svc *Service, appURL string) *Handler {
	return &Handler{svc: svc, appURL: appURL}
}

// RegisterRoutes mounts all auth endpoints on r.
func (h *Handler) RegisterRoutes(r gin.IRouter, jwtAuth gin.HandlerFunc, loginLimiter gin.HandlerFunc, registerLimiter gin.HandlerFunc) {
	r.POST("/register", registerLimiter, h.register)
	r.POST("/verify-code", registerLimiter, h.verifyCode)
	r.POST("/login", loginLimiter, h.login)
	r.GET("/me", jwtAuth, h.me)
}

// register handles POST /auth/register: creates an unverified account and sends a 6-digit OTP.
func (h *Handler) register(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Register(c.Request.Context(), body.Email, body.Password); err != nil {
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

// verifyCode handles POST /auth/verify-code: checks the OTP and returns a JWT on success.
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

// login handles POST /auth/login: verifies credentials and returns a signed JWT on success.
func (h *Handler) login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required"`
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

// me handles GET /auth/me: returns the profile of the authenticated user.
func (h *Handler) me(c *gin.Context) {
	userID := c.GetString(mw.UserIDKey)
	user, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, user)
}
