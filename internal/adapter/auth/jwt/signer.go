package jwt

import (
	"errors"
	"time"

	"pvz-service/internal/domain/user"
	"pvz-service/internal/usecase/ports"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManagerJWT struct {
	secret []byte
}

func NewTokenManagerJWT(secret string) ports.TokenManager {
	return &TokenManagerJWT{secret: []byte(secret)}
}

func (tm *TokenManagerJWT) GenerateToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"role":    u.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secret)
}

func (tm *TokenManagerJWT) ParseToken(tokenStr string) (*user.User, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}
		return tm.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	uid, _ := claims["user_id"].(string)
	role, _ := claims["role"].(string)
	return &user.User{ID: uid, Role: role}, nil
}
