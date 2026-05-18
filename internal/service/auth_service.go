package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"weicloud-backend/internal/model"
)

type AuthService struct {
	db             *gorm.DB
	jwtSecret      []byte
	jwtExpireHours int
}

type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var ErrInvalidCredential = errors.New("invalid username or password")
var ErrUserDisabled = errors.New("user is disabled")

func NewAuthService(db *gorm.DB, jwtSecret string, jwtExpireHours int) *AuthService {
	return &AuthService{
		db:             db,
		jwtSecret:      []byte(jwtSecret),
		jwtExpireHours: jwtExpireHours,
	}
}

func (s *AuthService) Login(username, password string) (model.User, string, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, "", ErrInvalidCredential
		}
		return model.User{}, "", fmt.Errorf("find user by username: %w", err)
	}

	if user.Status != model.UserStatusActive {
		return model.User{}, "", ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.User{}, "", ErrInvalidCredential
	}

	token, err := s.signToken(user)
	if err != nil {
		return model.User{}, "", err
	}
	return user, token, nil
}

func (s *AuthService) signToken(user model.User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.jwtExpireHours) * time.Hour)

	claims := UserClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign jwt token: %w", err)
	}
	return signed, nil
}

func (s *AuthService) ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
