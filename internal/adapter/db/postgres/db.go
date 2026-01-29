package postgres

import (
	"context"
	"fmt"

	"pvz-service/internal/adapter/db/repo"
	"pvz-service/internal/usecase/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

func (db *PostgresDB) Close() {
	db.pool.Close()
}

func NewDB(host string, port int, user, password, dbName string) (*PostgresDB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbName)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{pool: pool}, nil
}

func (db *PostgresDB) Ping(ctx context.Context) error {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return conn.Ping(ctx)
}

func (db *PostgresDB) Begin(ctx context.Context) (ports.Tx, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &PostgresTx{
		tx:            tx,
		userRepo:      repo.NewUserRepo(tx),
		pvzRepo:       repo.NewPVZRepo(tx),
		receptionRepo: repo.NewReceptionRepo(tx),
		prodRepo:      repo.NewProductRepo(tx),
	}, nil
}

func (db *PostgresDB) UserRepo() ports.UserRepository {
	return repo.NewUserRepo(db.pool)
}
func (db *PostgresDB) PVZRepo() ports.PVZRepository {
	return repo.NewPVZRepo(db.pool)
}
func (db *PostgresDB) ReceptionRepo() ports.ReceptionRepository {
	return repo.NewReceptionRepo(db.pool)
}
func (db *PostgresDB) ProductRepo() ports.ProductRepository {
	return repo.NewProductRepo(db.pool)
}
