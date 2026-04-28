package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huimingz/hyweb-assessment/model"
	"github.com/huimingz/hyweb-assessment/service"
)

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterRequest  true  "Register payload"
// @Success      200   {object}  Response
// @Failure      400   {object}  Response
// @Failure      409   {object}  Response
// @Failure      500   {object}  Response
// @Router       /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			Fail(c, http.StatusConflict, "email already exists", nil)
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "register failed", "error", err)
		Fail(c, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	OK(c, "registered successfully", gin.H{
		"email":   user.Email,
		"created": user.Created.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Login godoc
// @Summary      Login and obtain a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.LoginRequest  true  "Login payload"
// @Success      200   {object}  Response
// @Failure      400   {object}  Response
// @Failure      401   {object}  Response
// @Failure      500   {object}  Response
// @Router       /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	user, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			Fail(c, http.StatusUnauthorized, "invalid credentials", nil)
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "login failed", "error", err)
		Fail(c, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	token, err := generateToken(h.cfg.JWTSecret, user)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "token generation failed", "error", err, slog.String("email", user.Email))
		Fail(c, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	OK(c, "login successful", gin.H{"token": token})
}
