package repo

import (
	"context"
	"fmt"
	"time"

	"pvz-service/internal/domain/reception"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresReceptionRepo struct {
	conn interface {
		Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
		Query(context.Context, string, ...interface{}) (pgx.Rows, error)
		QueryRow(context.Context, string, ...interface{}) pgx.Row
	}
}

func NewReceptionRepo(conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}) *PostgresReceptionRepo {
	return &PostgresReceptionRepo{conn: conn}
}

func (r *PostgresReceptionRepo) Create(ctx context.Context, rec *reception.Reception) error {
	_, err := r.conn.Exec(ctx,
		"INSERT INTO receptions(id, pvz_id, started_at, status) VALUES($1,$2,$3,$4)",
		rec.ID, rec.PVZID, rec.StartedAt, rec.Status)
	return err
}

func (r *PostgresReceptionRepo) GetOpenByPVZ(ctx context.Context, pvzID string) (*reception.Reception, error) {
	row := r.conn.QueryRow(ctx,
		"SELECT id, pvz_id, started_at, status FROM receptions WHERE pvz_id=$1 AND status=$2",
		pvzID, reception.StatusInProgress)
	var rec reception.Reception
	err := row.Scan(&rec.ID, &rec.PVZID, &rec.StartedAt, &rec.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *PostgresReceptionRepo) Close(ctx context.Context, receptionID string) error {
	_, err := r.conn.Exec(ctx,
		"UPDATE receptions SET status=$1 WHERE id=$2", reception.StatusClosed, receptionID)
	return err
}

func (r *PostgresReceptionRepo) GetByPVZ(ctx context.Context, pvzID string, from, to *time.Time) ([]reception.Reception, error) {
	query := "SELECT id, pvz_id, started_at, status FROM receptions WHERE pvz_id=$1"
	args := []any{pvzID}
	if from != nil {
		args = append(args, *from)
		query += fmt.Sprintf(" AND started_at >= $%d", len(args))
	}
	if to != nil {
		args = append(args, *to)
		query += fmt.Sprintf(" AND started_at <= $%d", len(args))
	}
	query += " ORDER BY started_at"
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var recs []reception.Reception
	for rows.Next() {
		var rec reception.Reception
		if err := rows.Scan(&rec.ID, &rec.PVZID, &rec.StartedAt, &rec.Status); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return recs, nil
}
