package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
	"log"
	"sync"
	"time"
)

type UserRepository struct {
	Conn *pgx.Conn
	Mu   sync.Mutex
	Ctx  context.Context
}

func NewUserRepository(ctx context.Context, dbURL string) (*UserRepository, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctxTimeOut, dbURL)
	if err != nil {
		return nil, fmt.Errorf("NewUserRepository() connect: %w", err)
	}
	var mtx sync.Mutex
	return &UserRepository{Conn: conn, Mu: mtx, Ctx: ctx}, nil
}

func (r *UserRepository) Close() {
	err := r.Conn.Close(context.Background())
	if err != nil {
		log.Printf("Close(): %v", err)
	}
}

func (r *UserRepository) RepoInsertUser(ctx context.Context, user models.User) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	err := r.Conn.QueryRow(ctxWithDeadline, "INSERT INTO users (login, full_name) VALUES ($1, $2) RETURNING *", user.Login, user.FullName).
		Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("RepoInsertUser() QueryRow: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) RepoGetUserById(ctx context.Context, id int) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	var user models.User
	err := r.Conn.QueryRow(ctxWithDeadline, "SELECT id, login, full_name, created_at FROM users WHERE id=$1", id).
		Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("RepoGetUserById() QueryRow:  %w", err)
	}
	return &user, err
}

func (r *UserRepository) RepoGetUsersList(ctx context.Context, sqlQuery string) (*[]models.User, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := r.Conn.Query(ctxTimeOut, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("RepoGetUsersList() Query:%w", err)
	}

	var users []models.User
	for rows.Next() {
		user := *new(models.User)
		err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("RepoGetUsersList() Scan:%w", err)
		}
		users = append(users, user)
	}
	return &users, nil

}

func (r *UserRepository) RepoCreatePost(ctx context.Context, post models.Post) (*models.Post, error) {

	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var userId int
	err := r.Conn.QueryRow(ctxTimeOut, "SELECT id FROM users WHERE id=$1", post.UserId).Scan(&userId)
	if err != nil {
		return nil, fmt.Errorf("RepoCreatePost() QueryRow(SELECT): %w", err)
	}

	err = r.Conn.QueryRow(ctxTimeOut, "INSERT INTO posts (user_id,text) VALUES ($1,$2) RETURNING *", post.UserId, post.Text).
		Scan(&post.ID, &post.UserId, &post.Text, &post.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("RepoCreatePost() QueryRow(INSERT) %w", err)
	}
	return &post, nil
}

func (r *UserRepository) RepoGetAllPostsUser(ctx context.Context, sqlQuery string) (*[]models.Post, error) {

	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := r.Conn.Query(ctxTimeOut, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("RepoGetAllPostsUser() Query:%w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err = rows.Scan(&post.ID, &post.UserId, &post.Text, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("RepoGetAllPostsUser() Scan:%w", err)
		}
		posts = append(posts, post)
	}
	return &posts, nil
}

func (r *UserRepository) RepoGetAllUsers(ctx context.Context) (*[]models.User, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := r.Conn.Query(ctxTimeOut, "SELECT *FROM users")
	if err != nil {

		return nil, fmt.Errorf("RepoGetAllUsers() Query: %w", err)
	}

	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("RepoGetAllUsers() Scan: %w", err)
		}
		users = append(users, user)
	}
	return &users, nil
}
