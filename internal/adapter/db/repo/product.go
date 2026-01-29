package repo

import (
	"context"

	"pvz-service/internal/domain/product"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresProductRepo struct {
	conn interface {
		Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
		QueryRow(context.Context, string, ...interface{}) pgx.Row
		Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	}
}

func NewProductRepo(conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}) *PostgresProductRepo {
	return &PostgresProductRepo{conn: conn}
}

func (r *PostgresProductRepo) Create(ctx context.Context, pr *product.Product) error {
	_, err := r.conn.Exec(ctx,
		"INSERT INTO products(id, reception_id, added_at, type) VALUES($1,$2,$3,$4)",
		pr.ID, pr.ReceptionID, pr.AddedAt, pr.Type)
	return err
}

func (r *PostgresProductRepo) GetLastByReception(ctx context.Context, receptionID string) (*product.Product, error) {
	row := r.conn.QueryRow(ctx,
		"SELECT id, reception_id, added_at, type FROM products WHERE reception_id=$1 ORDER BY added_at DESC LIMIT 1",
		receptionID)
	var pr product.Product
	err := row.Scan(&pr.ID, &pr.ReceptionID, &pr.AddedAt, &pr.Type)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

func (r *PostgresProductRepo) Delete(ctx context.Context, productID string) error {
	_, err := r.conn.Exec(ctx,
		"DELETE FROM products WHERE id=$1", productID)
	return err
}

func (r *PostgresProductRepo) GetByReception(ctx context.Context, receptionID string) ([]product.Product, error) {
	rows, err := r.conn.Query(ctx,
		"SELECT id, reception_id, added_at, type FROM products WHERE reception_id=$1 ORDER BY added_at",
		receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []product.Product
	for rows.Next() {
		var pr product.Product
		if err := rows.Scan(&pr.ID, &pr.ReceptionID, &pr.AddedAt, &pr.Type); err != nil {
			return nil, err
		}
		products = append(products, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
