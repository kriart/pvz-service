package repo

import (
	"context"

	"pvz-service/internal/domain/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresUserRepo struct {
	conn interface {
		Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
		QueryRow(context.Context, string, ...interface{}) pgx.Row
	}
}

func NewUserRepo(conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}) *PostgresUserRepo {
	return &PostgresUserRepo{conn: conn}
}

func (r *PostgresUserRepo) Create(ctx context.Context, u *user.User) error {
	_, err := r.conn.Exec(ctx,
		"INSERT INTO users(id, email, password_hash, role, created_at) VALUES($1,$2,$3,$4,$5)",
		u.ID, u.Email, u.PasswordHash, u.Role, u.CreatedAt)
	return err
}

func (r *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.conn.QueryRow(ctx,
		"SELECT id, email, password_hash, role, created_at FROM users WHERE email=$1", email)
	var u user.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
