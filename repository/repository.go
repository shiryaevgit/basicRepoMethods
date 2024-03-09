package repository

import (
	"context"
	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
)

type UserInterface interface {
	Close()
	CreateUser(ctx context.Context, user models.User) (*models.User, error)
	GetUserById(ctx context.Context, userId int) (*models.User, error)
	GetUsersList(ctx context.Context, login, orderBy, limit, offset string) (*[]models.User, error)
	CreatePost(ctx context.Context, post models.Post) (*models.Post, error)
	GetAllPostsUser(ctx context.Context, userId, limit, offset string) (*[]models.Post, error)
	GetAllUsers(ctx context.Context) (*[]models.User, error)
	CheckUser(ctx context.Context, userId int) error
}
