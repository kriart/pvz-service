package ports

import "context"

type Tx interface {
	UserRepo() UserRepository
	PVZRepo() PVZRepository
	ReceptionRepo() ReceptionRepository
	ProductRepo() ProductRepository
	Commit() error
	Rollback() error
}

type TxManager interface {
	Begin(ctx context.Context) (Tx, error)
}
