package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"

	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
)

type RepoPostgres struct {
	Ctx  context.Context // контекст в структуре - антипаттерн
	Mu   sync.Mutex      // делаем поля неэкспортируемыми (с маленькой буквы)
	conn *pgx.Conn       // используем pgxpool заместо conn
}

func NewRepoPostgres(ctx context.Context, dbURL string) (*RepoPostgres, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Можно сдвинуть в main ко всем инъекциям и передавать входным параметром в конструктор NewRepoPostgres
	conn, err := pgx.Connect(connectCtx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("NewRepoPostgres() connect: %w", err)
	}

	var mtx sync.Mutex // зачем мьютекс?

	return &RepoPostgres{conn: conn, Mu: mtx, Ctx: ctx}, nil
}

// передавай контекст в close, возвращай ошибку наверх
func (r *RepoPostgres) Close() {
	err := r.conn.Close(context.Background())
	if err != nil {
		log.Printf("Close(): %v", err)
	}
}

func (r *RepoPostgres) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel() // лучше context.WithTimeout

	sqlQuery, _, err := goqu.Insert("users").
		Cols("login", "full_name").
		Vals(goqu.Vals{user.Login, user.FullName}).
		Returning("*"). // не используем *
		ToSQL()

	if err != nil {
		return nil, fmt.Errorf("GetUserById() ToSQL:  %w", err)
	}

	err = r.conn.QueryRow(ctxWithDeadline, sqlQuery).
		Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
	if err != nil {
		// добавь constraint на уникальность login, обработай ошибку
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
	err = r.conn.QueryRow(ctxWithDeadline, sqlQuery).
		Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
	if err != nil {
		// обработай ошибку pgx.ErrNoRows
		// pgx.ErrNoRows
		return nil, fmt.Errorf("GetUserById() QueryRow:  %w", err)
	}
	return &user, err
}

// поинтер на слайс в возвращающемся аргументе необоснован
// выполни конвертацию limit, offset в uint в слое выше
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
			// Когда перенесешь в хэндлер, верни пользователю статус код 400
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

	rows, err := r.conn.Query(ctxTimeOut, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("GetUsersList() Query:%w", err)
	}

	users := make([]models.User, 0, 100)

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetUsersList() Scan:%w", err)
		}
		users = append(users, user)
	}

	return &users, nil
}

func (r *RepoPostgres) CreatePost(ctx context.Context, post models.Post) (*models.Post, error) {
	// затемняем родительский контекст, так как все равно не будем к нему обращаться
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	sqlQuery, _, err := goqu.Insert("posts").
		Cols("user_id", "text").
		Vals(goqu.Vals{post.UserId, post.Text}).
		Returning("*"). // лучше перечислить колонки а не использовать *
		ToSQL()

	if err != nil {
		return nil, fmt.Errorf("CreatePost() ToSQL:  %w", err)
	}

	err = r.conn.QueryRow(ctx, sqlQuery).
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
		// limit в хэндлер
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

	rows, err := r.conn.Query(ctxTimeOut, sqlQuery)
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

	rows, err := r.conn.Query(ctxTimeOut, sqlQuery)
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
	err = r.conn.QueryRow(ctxTimeOut, sqlQueryCheck).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("CheckUser(): user with ID:%d not found", userId)
		}
		return fmt.Errorf("CheckUser() QueryRow(SELECT): %w", err)
	}
	return nil
}
