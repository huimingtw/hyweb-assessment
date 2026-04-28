package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserService(db *sql.DB, logger *slog.Logger) *UserService {
	return &UserService{db: db, logger: logger}
}

func (s *UserService) ChangePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	email = strings.ToLower(email)

	var currentHash string
	err := s.db.QueryRowContext(ctx,
		"SELECT password FROM users WHERE email = ?", email,
	).Scan(&currentHash)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "change password: db query failed", "error", err, "email", email)
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword)); err != nil {
		return ErrWrongPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET password = ? WHERE email = ?",
		string(newHash), email,
	)
	if err != nil {
		s.logger.ErrorContext(ctx, "change password: db update failed", "error", err, "email", email)
		return err
	}

	s.logger.DebugContext(ctx, "password changed", "email", email)
	return nil
}
