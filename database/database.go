package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/shiryaevgit/myProject/pkg/models"
)

type UserRepository struct {
	Сonn *pgx.Conn
}

func (db *UserRepository) Close() {
	db.Сonn.Close(context.Background()) //
}
func NewHandlerDB(dbURL string) (*UserRepository, error) {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect(): %w", err)
	}
	return &UserRepository{conn}, nil
}

func (db *UserRepository) GetAllUsers() ([]models.User, error) {
	rows, err := db.Сonn.Query(context.Background(), "SELECT * FROM users")
	if err != nil {
		return nil, fmt.Errorf("query(): %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan(): %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (db *UserRepository) GetAllPosts(userID int) ([]models.Post, error) {
	rows, err := db.Сonn.Query(context.Background(), "SELECT * FROM posts WHERE id=$1", userID)
	if err != nil {
		return nil, fmt.Errorf("query(): %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err = rows.Scan(&post.ID, &post.Text, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan(): %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (db *UserRepository) InsertIntoTestTable(login string, fullName string) error {
	_, err := db.Сonn.Exec(context.Background(), "INSERT INTO users (login,full_name) VALUES ($1, $2)", login, fullName)
	if err != nil {
		return fmt.Errorf("exec(): %w", err)
	}
	return nil
}
