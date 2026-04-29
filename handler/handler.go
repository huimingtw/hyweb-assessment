package handler

import (
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/huimingz/hyweb-assessment/config"
	"github.com/huimingz/hyweb-assessment/model"
	"github.com/huimingz/hyweb-assessment/service"
)

type Handler struct {
	authSvc    *service.AuthService
	userSvc    *service.UserService
	weatherSvc *service.WeatherService
	cfg        *config.Config
	logger     *slog.Logger
}

func NewHandler(
	authSvc *service.AuthService,
	userSvc *service.UserService,
	weatherSvc *service.WeatherService,
	cfg *config.Config,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		authSvc:    authSvc,
		userSvc:    userSvc,
		weatherSvc: weatherSvc,
		cfg:        cfg,
		logger:     logger,
	}
}

func generateToken(secret string, user *model.User) (string, error) {
	claims := model.JWTClaims{
		Email:   user.Email,
		Updated: user.Updated.UTC().Format(time.RFC3339),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
