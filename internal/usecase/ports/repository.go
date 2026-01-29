package ports

import (
	"context"
	"pvz-service/internal/domain/product"
	"pvz-service/internal/domain/pvz"
	"pvz-service/internal/domain/reception"
	"pvz-service/internal/domain/user"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	FindByEmail(ctx context.Context, email string) (*user.User, error)
}

type PVZRepository interface {
	Create(ctx context.Context, p *pvz.PVZ) error
	Get(ctx context.Context, id string) (*pvz.PVZ, error)
	List(ctx context.Context, from, to *time.Time, limit, offset int) ([]pvz.PVZ, error)
}

type ReceptionRepository interface {
	Create(ctx context.Context, r *reception.Reception) error
	GetOpenByPVZ(ctx context.Context, pvzID string) (*reception.Reception, error)
	Close(ctx context.Context, receptionID string) error
	GetByPVZ(ctx context.Context, pvzID string, from, to *time.Time) ([]reception.Reception, error)
}

type ProductRepository interface {
	Create(ctx context.Context, pr *product.Product) error
	GetLastByReception(ctx context.Context, receptionID string) (*product.Product, error)
	Delete(ctx context.Context, productID string) error
	GetByReception(ctx context.Context, receptionID string) ([]product.Product, error)
}
