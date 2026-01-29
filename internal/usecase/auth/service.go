package auth

import (
	"context"
	"strings"

	"pvz-service/internal/domain/user"
	"pvz-service/internal/usecase/ports"

	"github.com/google/uuid"
)

type Service struct {
	userRepo       ports.UserRepository
	tokenManager   ports.TokenManager
	passwordHasher ports.PasswordHasher
	clock          ports.Clock
}

func NewService(userRepo ports.UserRepository, tokenManager ports.TokenManager, passwordHasher ports.PasswordHasher, clock ports.Clock) *Service {
	return &Service{userRepo: userRepo, tokenManager: tokenManager, passwordHasher: passwordHasher, clock: clock}
}

func (s *Service) DummyLogin(ctx context.Context, userType string) (*AuthToken, error) {
	userType = strings.ToLower(userType)
	if userType != user.RoleClient && userType != user.RoleModerator {
		return nil, ErrInvalidUserType
	}
	dummyUser := &user.User{
		ID:   uuid.New().String(),
		Role: userType,
	}
	token, err := s.tokenManager.GenerateToken(dummyUser)
	if err != nil {
		return nil, err
	}
	return &AuthToken{Token: token}, nil
}

func (s *Service) Register(ctx context.Context, email, password, userType string) (*RegisterResult, error) {
	email = strings.ToLower(email)
	userType = strings.ToLower(userType)
	if userType != user.RoleClient && userType != user.RoleModerator {
		return nil, ErrInvalidUserType
	}
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, user.ErrEmailAlreadyExists
	}
	hash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, err
	}
	uid := uuid.New().String()
	u := &user.User{
		ID:           uid,
		Email:        email,
		PasswordHash: hash,
		Role:         userType,
		CreatedAt:    s.clock.Now(),
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	return &RegisterResult{UserID: uid}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthToken, error) {
	email = strings.ToLower(email)
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, user.ErrInvalidCredentials
	}
	match := s.passwordHasher.Compare(u.PasswordHash, password)
	if !match {
		return nil, user.ErrInvalidCredentials
	}
	token, err := s.tokenManager.GenerateToken(u)
	if err != nil {
		return nil, err
	}
	return &AuthToken{Token: token}, nil
}
