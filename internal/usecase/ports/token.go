package ports

import "pvz-service/internal/domain/user"

type TokenManager interface {
	GenerateToken(u *user.User) (string, error)
	ParseToken(tokenStr string) (*user.User, error)
}
