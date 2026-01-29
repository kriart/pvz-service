package postgres

import (
	"context"

	"pvz-service/internal/usecase/ports"

	"github.com/jackc/pgx/v5"
)

type PostgresTx struct {
	tx            pgx.Tx
	userRepo      ports.UserRepository
	pvzRepo       ports.PVZRepository
	receptionRepo ports.ReceptionRepository
	prodRepo      ports.ProductRepository
}

func (t *PostgresTx) UserRepo() ports.UserRepository {
	return t.userRepo
}
func (t *PostgresTx) PVZRepo() ports.PVZRepository {
	return t.pvzRepo
}
func (t *PostgresTx) ReceptionRepo() ports.ReceptionRepository {
	return t.receptionRepo
}
func (t *PostgresTx) ProductRepo() ports.ProductRepository {
	return t.prodRepo
}
func (t *PostgresTx) Commit() error {
	return t.tx.Commit(context.Background())
}
func (t *PostgresTx) Rollback() error {
	return t.tx.Rollback(context.Background())
}
