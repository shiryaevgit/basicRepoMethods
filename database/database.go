package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"sync"
)

type UserRepository struct {
	Conn *pgx.Conn
	Mu   sync.Mutex
	Ctx  context.Context
}

func NewUserRepository(dbURL string, ctx context.Context) (*UserRepository, error) {
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect(): %w", err)
	}

	var mtx sync.Mutex

	return &UserRepository{Conn: conn, Mu: mtx, Ctx: ctx}, nil
}
func (db *UserRepository) Close() {
	db.Conn.Close(context.Background()) //
}
