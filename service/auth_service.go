package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/huimingz/hyweb-assessment/model"
	"golang.org/x/crypto/bcrypt"
)

const BcryptCost = 12

type AuthService struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAuthService(db *sql.DB, logger *slog.Logger) *AuthService {
	return &AuthService{db: db, logger: logger}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	email = strings.ToLower(email)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO users (email, password) VALUES (?, ?)",
		email, string(hash),
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrEmailExists
		}
		s.logger.ErrorContext(ctx, "register: db insert failed", "error", err)
		return nil, err
	}

	user := &model.User{
		Email:   email,
		Created: time.Now().UTC(),
		Updated: time.Now().UTC(),
	}

	// Re-query to get exact DB-generated timestamps
	row := s.db.QueryRowContext(ctx,
		"SELECT created, updated FROM users WHERE email = ?", email,
	)
	if scanErr := row.Scan(&user.Created, &user.Updated); scanErr != nil {
		s.logger.WarnContext(ctx, "register: failed to re-query timestamps", "error", scanErr)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, error) {
	email = strings.ToLower(email)

	user := &model.User{}
	err := s.db.QueryRowContext(ctx,
		"SELECT email, password, created, updated FROM users WHERE email = ?", email,
	).Scan(&user.Email, &user.Password, &user.Created, &user.Updated)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "login: db query failed", "error", err)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
