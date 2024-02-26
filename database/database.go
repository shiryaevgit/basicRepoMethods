package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

type DataBaseHandler struct {
	conn *pgx.Conn
}

func (db *DataBaseHandler) Close() {
	db.conn.Close(context.Background()) //
}

func NewHandlerDB(dbURL string) (*DataBaseHandler, error) {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect(): %w", err)
	}
	return &DataBaseHandler{conn}, nil
}

func (db *DataBaseHandler) SelectFromTestTable() error {
	rows, err := db.conn.Query(context.Background(), "SELECT * FROM users")
	if err != nil {
		return fmt.Errorf("query(): %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var created_at time.Time
		var login string
		var full_name string

		if err = rows.Scan(&id, &created_at, &login, &full_name); err != nil {
			return fmt.Errorf("scan(): %w", err)
		}
		fmt.Printf("id: %d, login: %s, full_name: %s, created_at:%v\n ", id, created_at, login, full_name)
	}
	return nil
}

func (db *DataBaseHandler) InsertIntoTestTable(login string, fullName string) error {
	_, err := db.conn.Exec(context.Background(), "INSERT INTO users (login,full_name) VALUES ($1, $2)", login, fullName)
	if err != nil {
		return fmt.Errorf("exec(): %w", err)
	}
	return nil
}
