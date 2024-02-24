package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
)

type DataBaseHandler struct {
	conn *pgx.Conn
}

func (db *DataBaseHandler) Close() {
	db.conn.Close(context.Background()) //
}

func NewHandlerDB() (*DataBaseHandler, error) {
	// conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL")) - не работает, та же ошибка
	conn, err := pgx.Connect(context.Background(), "user=postgres dbname=postgres sslmode=disable password=postgres")
	if err != nil {
		return nil, fmt.Errorf("connect(): %v", err)
	}
	return &DataBaseHandler{conn}, nil
}

func (db *DataBaseHandler) SelectFromTestTable() error {
	rows, err := db.conn.Query(context.Background(), "SELECT * FROM test_table")
	if err != nil {
		return fmt.Errorf("query(): %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var age int

		if err = rows.Scan(&id, &name, &age); err != nil {
			return fmt.Errorf("scan(): %v", err)
		}

		fmt.Printf("id: %d, name: %s, age: %d\n", id, name, age)
	}

	return nil
}

func (db *DataBaseHandler) InsertIntoTestTable(name string, age int) error {
	_, err := db.conn.Exec(context.Background(), "INSERT INTO test_table (name, age) VALUES ($1, $2)", name, age)
	if err != nil {
		return fmt.Errorf("exec(): %v", err)
	}
	return nil
}
