package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pvz-service/internal/domain/pvz"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresPVZRepo struct {
	conn interface {
		Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
		Query(context.Context, string, ...interface{}) (pgx.Rows, error)
		QueryRow(context.Context, string, ...interface{}) pgx.Row
	}
}

func NewPVZRepo(conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}) *PostgresPVZRepo {
	return &PostgresPVZRepo{conn: conn}
}

func (r *PostgresPVZRepo) Create(ctx context.Context, p *pvz.PVZ) error {
	_, err := r.conn.Exec(ctx,
		"INSERT INTO pvzs(id, city, created_at) VALUES($1,$2,$3)",
		p.ID, p.City, p.CreatedAt)
	return err
}

func (r *PostgresPVZRepo) Get(ctx context.Context, id string) (*pvz.PVZ, error) {
	row := r.conn.QueryRow(ctx,
		"SELECT id, city, created_at FROM pvzs WHERE id=$1", id)
	var pv pvz.PVZ
	err := row.Scan(&pv.ID, &pv.City, &pv.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &pv, nil
}

func (r *PostgresPVZRepo) List(ctx context.Context, from, to *time.Time, limit, offset int) ([]pvz.PVZ, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if from == nil && to == nil {
		query := "SELECT p.id, p.city, p.created_at FROM pvzs p ORDER BY p.created_at"
		args := []any{}
		if limit > 0 {
			args = append(args, limit, offset)
			query = fmt.Sprintf("%s LIMIT $1 OFFSET $2", query)
		}
		rows, err = r.conn.Query(ctx, query, args...)
	} else {
		query := "SELECT DISTINCT p.id, p.city, p.created_at FROM pvzs p JOIN receptions r ON p.id = r.pvz_id"
		args := []any{}
		whereParts := []string{}
		if from != nil {
			args = append(args, *from)
			whereParts = append(whereParts, fmt.Sprintf("r.started_at >= $%d", len(args)))
		}
		if to != nil {
			args = append(args, *to)
			whereParts = append(whereParts, fmt.Sprintf("r.started_at <= $%d", len(args)))
		}
		if len(whereParts) > 0 {
			query += " WHERE " + strings.Join(whereParts, " AND ")
		}
		query += " GROUP BY p.id, p.city, p.created_at ORDER BY p.created_at"
		if limit > 0 {
			args = append(args, limit, offset)
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
		}
		rows, err = r.conn.Query(ctx, query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []pvz.PVZ{}
	for rows.Next() {
		var pv pvz.PVZ
		if err := rows.Scan(&pv.ID, &pv.City, &pv.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, pv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
