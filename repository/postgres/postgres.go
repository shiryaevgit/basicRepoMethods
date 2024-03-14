package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/shiryaevgit/basicRepoMethods/pkg/models"

	"log"
	"strconv"
	"sync"
	"time"
)

type RepoPostgres struct {
	Ctx  context.Context
	Mu   sync.Mutex
	Conn *pgx.Conn
}

func NewRepoPostgres(terminateContext context.Context, dbURL string) (*RepoPostgres, error) {
	ctxTimeOut, cancel := context.WithTimeout(terminateContext, 3*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctxTimeOut, dbURL)
	if err != nil {
		return nil, fmt.Errorf("NewRepoPostgres() connect: %w", err)
	}
	var mtx sync.Mutex
	return &RepoPostgres{Conn: conn, Mu: mtx, Ctx: terminateContext}, nil
}

func (r *RepoPostgres) Close() {
	err := r.Conn.Close(context.Background())
	if err != nil {
		log.Printf("Close(): %v", err)
	}
}

func (r *RepoPostgres) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	sqlQuery, _, err := goqu.Insert("users").
		Cols("login", "full_name").
		Vals(goqu.Vals{user.Login, user.FullName}).
		Returning("*").
		ToSQL()

	if err != nil {
		return nil, fmt.Errorf("GetUserById() ToSQL:  %w", err)
	}

	err = r.Conn.QueryRow(ctxWithDeadline, sqlQuery).
		Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf(" CreateUser() QueryRow: %w", err)
	}
	return &user, nil
}

func (r *RepoPostgres) GetUserById(ctx context.Context, userId int) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	sqlQuery, _, err := goqu.From("users").
		Select("id", "login", "full_name", "created_at").
		Where(goqu.Ex{"id": userId}).
		ToSQL()

	if err != nil {
		return nil, fmt.Errorf("GetUserById() ToSQL:  %w", err)
	}

	var user models.User
	err = r.Conn.QueryRow(ctxWithDeadline, sqlQuery).
		Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("GetUserById() QueryRow:  %w", err)
	}
	return &user, err
}

func (r *RepoPostgres) GetUsersList(ctx context.Context, login, orderBy, limit, offset string) (*[]models.User, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	ds := goqu.From("users")

	if login != "" {
		ds = ds.Where(goqu.C("login").Eq(login))
	}
	if orderBy != "" {
		ds = ds.Order(goqu.I(orderBy).Asc())
	}
	if limit != "" {
		limitInt, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetUsersList() ParseUint(limit): %w", err)
		}
		ds = ds.Limit(uint(limitInt))
	}
	if offset != "" {
		offsetInt, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetUsersList() ParseUint(offset): %w", err)
		}
		ds = ds.Offset(uint(offsetInt))
	}

	sqlQuery, _, err := ds.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("GetUsersList() ToSQL: %w", err)
	}

	rows, err := r.Conn.Query(ctxTimeOut, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("GetUsersList() Query:%w", err)
	}

	users := make([]models.User, 0, 100)

	for rows.Next() {
		user := *new(models.User)
		err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetUsersList() Scan:%w", err)
		}
		users = append(users, user)
	}
	return &users, nil

}

func (r *RepoPostgres) CreatePost(ctx context.Context, post models.Post) (*models.Post, error) {

	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlQuery, _, err := goqu.Insert("posts").
		Cols("user_id", "text").
		Vals(goqu.Vals{post.UserId, post.Text}).
		Returning("*").
		ToSQL()

	if err != nil {
		return nil, fmt.Errorf("CreatePost() ToSQL:  %w", err)
	}

	err = r.Conn.QueryRow(ctxTimeOut, sqlQuery).
		Scan(&post.ID, &post.UserId, &post.Text, &post.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("CreatePost() QueryRow() %w", err)
	}
	return &post, nil
}

func (r *RepoPostgres) GetAllPostsUser(ctx context.Context, userId, limit, offset string) (*[]models.Post, error) {

	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	ds := goqu.From("posts")

	if userId != "" {
		ds = ds.Where(goqu.C("user_id").Eq(userId))
	}
	if limit != "" {
		limitInt, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetAllUsers() ParseUint(limit): %w", err)
		}
		ds = ds.Limit(uint(limitInt))
	}

	if offset != "" {
		offsetInt, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetAllUsers() ParseUint(offset): %w", err)
		}
		ds = ds.Offset(uint(offsetInt))
	}
	sqlQuery, _, _ := ds.ToSQL()

	rows, err := r.Conn.Query(ctxTimeOut, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("GetAllPostsUser() Query:%w", err)
	}
	defer rows.Close()

	posts := make([]models.Post, 0, 100)

	for rows.Next() {
		var post models.Post
		err = rows.Scan(&post.ID, &post.UserId, &post.Text, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetAllPostsUser() Scan:%w", err)
		}
		posts = append(posts, post)
	}
	return &posts, nil
}
func (r *RepoPostgres) GetAllUsers(ctx context.Context) (*[]models.User, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlQuery, _, _ := goqu.From("users").ToSQL()

	rows, err := r.Conn.Query(ctxTimeOut, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers() Query: %w", err)
	}

	users := make([]models.User, 0, 100)

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetAllUsers() Scan: %w", err)
		}
		users = append(users, user)
	}
	return &users, nil
}
func (r *RepoPostgres) CheckUser(ctx context.Context, userId int) error {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	sqlQueryCheck, _, err := goqu.Select("id").
		From("users").
		Where(goqu.Ex{"id": userId}).
		ToSQL()

	if err != nil {
		return fmt.Errorf("CheckUser() ToSQL:%w", err)
	}

	var id int
	err = r.Conn.QueryRow(ctxTimeOut, sqlQueryCheck).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("CheckUser(): user with ID:%d not found", userId)
		}
		return fmt.Errorf("CheckUser() QueryRow(SELECT): %w", err)
	}
	return nil
}
