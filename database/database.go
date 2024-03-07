package database

import (
	"context"
	"errors"
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

func NewUserRepository(terminateContext context.Context, dbURL string) (*UserRepository, error) {
	ctxTimeOut, cancel := context.WithTimeout(terminateContext, 1*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctxTimeOut, dbURL)
	if err != nil {
		return nil, fmt.Errorf("NewUserRepository() connect: %w", err)
	}
	var mtx sync.Mutex
	return &UserRepository{Conn: conn, Mu: mtx, Ctx: terminateContext}, nil
}

func (r *UserRepository) Close() {
	err := r.Conn.Close(context.Background())
	if err != nil {
		log.Printf("Close(): %v", err)
	}
}

func (r *UserRepository) RepoInsertUser(ctx context.Context, user models.User, sqlQuery string) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	err := r.Conn.QueryRow(ctxWithDeadline, sqlQuery, user.Login, user.FullName).
		Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("RepoInsertUser() QueryRow: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) RepoGetUserById(ctx context.Context, id int, sqlQuery string) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	var user models.User
	err := r.Conn.QueryRow(ctxWithDeadline, sqlQuery, id).
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

	users := make([]models.User, 0, 100)

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

func (r *UserRepository) RepoCreatePost(ctx context.Context, post models.Post, sqlQuery string) (*models.Post, error) {

	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := r.Conn.QueryRow(ctxTimeOut, sqlQuery, post.UserId, post.Text).
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

	posts := make([]models.Post, 0, 100)

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
func (r *UserRepository) RepoGetAllUsers(ctx context.Context, sqlQuery string) (*[]models.User, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := r.Conn.Query(ctxTimeOut, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("RepoGetAllUsers() Query: %w", err)
	}

	users := make([]models.User, 0, 100)

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
func (r *UserRepository) RepoCheckUser(ctx context.Context, userId int, sqlQuery string) error {
	var existingUserId int
	err := r.Conn.QueryRow(ctx, sqlQuery, userId).Scan(&existingUserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("RepoCheckUser(): user with ID:%d not found", userId)
		}
		return fmt.Errorf("RepoCheckUser() QueryRow(SELECT): %w", err)
	}
	return nil
}
