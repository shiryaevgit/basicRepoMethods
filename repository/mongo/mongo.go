package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
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

// назови контекст ctx
func NewRepoMongo(terminateContext context.Context, MongoURI string) (*RepoMongo, error) {
	ctxTimeOut, cancel := context.WithTimeout(terminateContext, 5*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(MongoURI)
	client, err := mongo.Connect(ctxTimeOut, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("NewRepoMongo() Connect: %w", err)
	}

	// ctxTimeOut используй тут
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("NewRepoMongo() Ping: %w", err)
	}

	// зачем мютекс?
	var mtx sync.Mutex

	//// Создаем базу данных и коллекции
	db := client.Database("mongo")
	users := db.Collection("users")
	posts := db.Collection("posts")

	return &RepoMongo{ctxTimeOut, mtx, db, Collections{users, posts}}, nil
}

func (r *RepoMongo) Close() {
	r.Mu.Lock() // зачем тут мьютекс?
	defer r.Mu.Unlock()

	if err := r.Conn.Client().Disconnect(r.Ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}
}

func (r *RepoMongo) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	ctxWithDeadline, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.users.InsertOne(ctxWithDeadline, user)
	if err != nil {
		return nil, fmt.Errorf("CreateUser() InsertOne:%w", err)
	}

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

	// лучше так:
	//if err := res.Err(); err != nil {
	//	if errors.Is(err, mongo.ErrNoDocuments) {
	//		return nil, fmt.Errorf("GetUserById() user with id: %d not found: %w", userId, err)
	//	}
	//	return nil, fmt.Errorf("GetUserById(): %w", err)
	//}

	var user models.User
	if err := res.Decode(&user); err != nil {
		return nil, fmt.Errorf("GetUserById() Decode: %w", err)
	}

	return &user, nil
}

func (r *RepoMongo) GetUsersList(ctx context.Context, login, orderBy, limit, offset string) (*[]models.User, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	filter := bson.M{}
	if login != "" {
		filter["login"] = login
	}

	findOptions := options.Find()
	if orderBy != "" {
		findOptions.SetSort(bson.D{{orderBy, 1}})
	}
	if limit != "" {
		limitInt, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetUsersList() ParseInt(limit): %w", err)
		}
		findOptions.SetLimit(limitInt)
	}
	if offset != "" {
		offsetInt, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetUsersList() ParseInt(offset): %w", err)
		}
		findOptions.SetSkip(offsetInt)
	}

	cursor, err := r.users.Find(ctxTimeOut, filter, findOptions) // запрос к коллекции
	if err != nil {
		return nil, fmt.Errorf("GetUsersList() Find: %w", err)
	}
	defer func() {
		err := cursor.Close(ctxTimeOut)
		if err != nil {
		}
	}() // обработай ошибку

	var users []models.User
	for cursor.Next(ctx) {
		var user models.User
		if err = cursor.Decode(&user); err != nil {
			return nil, fmt.Errorf("GetUsersList() Decode: %w", err)
		}
		users = append(users, user)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("GetUsersList() Cursor error: %w", err)
	}

	return &users, nil

}

func (r *RepoMongo) CreatePost(ctx context.Context, post models.Post) (*models.Post, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result, err := r.posts.InsertOne(ctxTimeOut, post)
	if err != nil {
		return nil, fmt.Errorf("CreatePost() InsertOne: %w", err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("CreatePost() InsertedID is not an ObjectID")
	}

	post.ID, err = strconv.Atoi(insertedID.Hex())
	if err != nil {
		return nil, fmt.Errorf("CreateUser() Atoi: %w", err)
	}

	return &post, nil
}

func (r *RepoMongo) GetAllPostsUser(ctx context.Context, userId, limit, offset string) (*[]models.Post, error) {

	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	filter := bson.M{}
	if userId != "" {
		filter["user_id"] = userId
	}

	options := options.Find()
	if limit != "" {
		limitInt, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetAllPostsUser() ParseInt(limit): %w", err)
		}
		options.SetLimit(limitInt)
	}
	if offset != "" {
		offsetInt, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GetAllPostsUser() ParseInt(offset): %w", err)
		}
		options.SetSkip(offsetInt)
	}

	cursor, err := r.posts.Find(ctxTimeOut, filter, options)
	if err != nil {
		return nil, fmt.Errorf("GetAllPostsUser() Find: %w", err)
	}
	defer cursor.Close(ctxTimeOut)

	var posts []models.Post
	for cursor.Next(ctxTimeOut) {
		var post models.Post
		if err = cursor.Decode(&post); err != nil {
			return nil, fmt.Errorf("GetAllPostsUser() Decode: %w", err)
		}
		posts = append(posts, post)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("GetAllPostsUser() Cursor error: %w", err)
	}

	return &posts, nil
}
func (r *RepoMongo) GetAllUsers(ctx context.Context) (*[]models.User, error) {
	ctxTimeOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	cursor, err := r.users.Find(ctxTimeOut, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers() Find: %w", err)
	}
	defer cursor.Close(ctxTimeOut)

	var users []models.User
	for cursor.Next(ctxTimeOut) {
		var user models.User
		if err = cursor.Decode(&user); err != nil {
			return nil, fmt.Errorf("GetAllUsers() Decode: %w", err)
		}
		users = append(users, user)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("GetAllUsers() Cursor error: %w", err)
	}

	return &users, nil
}
func (r *RepoMongo) CheckUser(ctx context.Context, userId int) error {
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
	defer cancel()

	filter := bson.D{{"id", userId}}

	res := r.users.FindOne(ctxWithDeadline, filter)

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return fmt.Errorf("CheckUser() user with id:%d not found", userId)
		}
	}
	return nil
}
