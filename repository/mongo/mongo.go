package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"sync"
	"time"
)

type RepoMongo struct {
	Ctx  context.Context
	Mu   sync.Mutex
	Conn *mongo.Database
	Collections
}

type Collections struct {
	users *mongo.Collection
	posts *mongo.Collection
}

func NewRepoMongo(terminateContext context.Context, MongoURI string) (*RepoMongo, error) {
	ctxTimeOut, cancel := context.WithTimeout(terminateContext, 5*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(MongoURI)
	client, err := mongo.Connect(ctxTimeOut, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("NewRepoMongo() Connect: %w", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("NewRepoMongo() Ping: %w", err)
	}

	var mtx sync.Mutex

	//// Создаем базу данных и коллекции
	db := client.Database("mongo")
	users := db.Collection("users")
	posts := db.Collection("posts")

	return &RepoMongo{ctxTimeOut, mtx, db, Collections{users, posts}}, nil
}

func (r *RepoMongo) Close() {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	// Закрываем соединение с MongoDB
	if err := r.Conn.Client().Disconnect(r.Ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}
}

func (r *RepoMongo) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	result, err := r.users.InsertOne(ctxWithDeadline, user)
	if err != nil {
		return nil, fmt.Errorf("CreateUser() InsertOne:%w", err)
	}

	// Получение идентификатора (ID) нового пользователя из результата вставки
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("CreateUser() Invalid inserted ID type")
	}

	user.ID, err = strconv.Atoi(insertedID.Hex())
	if err != nil {
		return nil, fmt.Errorf("CreateUser() Atoi: %w", err)
	}
	return &user, nil
}

func (r *RepoMongo) GetUserById(ctx context.Context, userId int) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	filter := bson.D{{"id", userId}}

	res := r.users.FindOne(ctxWithDeadline, filter)

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("GetUserById() user with id:%d not found", userId)
		}
		return nil, fmt.Errorf("GetUserById(): %w", res.Err())
	}
	var user models.User
	if err := res.Decode(&user); err != nil {
		return nil, fmt.Errorf("GetUserById() Decode: %w", err)
	}

	return &user, nil
}

func (r *RepoMongo) GetUsersList(ctx context.Context, login, orderBy, limit, offset string) (*[]models.User, error) {
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

func (r *RepoMongo) CreatePost(ctx context.Context, post models.Post) (*models.Post, error) {

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

func (r *RepoMongo) GetAllPostsUser(ctx context.Context, userId, limit, offset string) (*[]models.Post, error) {

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
func (r *RepoMongo) GetAllUsers(ctx context.Context) (*[]models.User, error) {
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
func (r *RepoMongo) CheckUser(ctx context.Context, userId int) error {

	sqlQueryCheck, _, err := goqu.Select("id").
		From("users").
		Where(goqu.Ex{"id": userId}).
		ToSQL()

	if err != nil {
		return fmt.Errorf("CheckUser() ToSQL:%w", err)
	}

	var id int
	err = r.Conn.QueryRow(ctx, sqlQueryCheck).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("CheckUser(): user with ID:%d not found", userId)
		}
		return fmt.Errorf("CheckUser() QueryRow(SELECT): %w", err)
	}
	return nil
}
