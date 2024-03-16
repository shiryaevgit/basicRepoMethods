package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
)

// Объявляем интерфейс в месте использования
type UserRepository interface {
	Close()
	CreateUser(ctx context.Context, user models.User) (*models.User, error)
	GetUserById(ctx context.Context, userId int) (*models.User, error)
	GetUsersList(ctx context.Context, login, orderBy, limit, offset string) (*[]models.User, error)
	CreatePost(ctx context.Context, post models.Post) (*models.Post, error)
	GetAllPostsUser(ctx context.Context, userId, limit, offset string) (*[]models.Post, error)
	GetAllUsers(ctx context.Context) (*[]models.User, error)
	CheckUser(ctx context.Context, userId int) error
}

type Handler struct {
	// Здесь мы должны зависеть от интерфейса а не от конкретной реализации
	userRepository UserRepository
}

func NewHandlerServ(repository UserRepository) *Handler {
	return &Handler{userRepository: repository}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("CreateUser(): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		/* Давай будем возвращать ответ в json формата:
		{
			"err": "ТЕКСТ ОШИБКИ"
		}
		*/

		// Не забывай делать return
	}

	// Используем контекст реквеста ( r.Context() )
	createdUser, err := h.userRepository.CreateUser(r.Context(), user)
	if err != nil {
		log.Printf("CreateUser(): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		// return ?
	}

	respJson, err := json.Marshal(createdUser)
	// пропущена обработка ошибки
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respJson)
	// Давай выводить ответ в JSON. Формат успешных ответов выбери самостоятельно
	if err != nil {
		log.Printf("CreateUser() Marshal: %v", err)
	}

}

func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	fmt.Println(idString) // используй логгер
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		log.Printf("GetUserById(): %v", err)
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	gotUser, err := h.userRepository.GetUserById(h.userRepository.Ctx, idInt)
	if err != nil {
		log.Printf("GetUserById(): %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	responseJSON, err := json.Marshal(gotUser)
	if err != nil {
		log.Printf("GetUserById(): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		// return?
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseJSON)
	if err != nil {
		log.Printf("GetUserById(): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) GetUsersList(w http.ResponseWriter, r *http.Request) {
	login := r.URL.Query().Get("login")
	orderBy := r.URL.Query().Get("orderBy")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	gotUsers, err := h.userRepository.GetUsersList(h.userRepository.Ctx, login, orderBy, limit, offset)
	if err != nil {
		log.Printf("GetUsersList() : %v", err)
		http.Error(w, "The entered data is incorrect", http.StatusBadRequest)
	}

	respJson, err := json.Marshal(gotUsers)
	if err != nil {
		log.Printf("GetUsersList() Marshal: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respJson)
	if err != nil {
		log.Printf("GetUsersList() Write: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {

	post := new(models.Post)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(post)
	if err != nil {
		log.Printf("CreatePost(): %v", err)
		http.Error(w, `Invalid request entered`, http.StatusBadRequest)
		return
	}

	if err = h.userRepository.CheckUser(h.userRepository.Ctx, post.UserId); err != nil {
		log.Printf("CreatePost(): %v", err)
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	createdPost, err := h.userRepository.CreatePost(h.userRepository.Ctx, *post)
	if err != nil {
		log.Printf("CreatePost(): %v", err)
		http.Error(w, "User not found", http.StatusBadRequest)
	}

	respJson, err := json.Marshal(createdPost)
	if err != nil {
		log.Printf("CreatePost() Marshal: %v", err)
		http.Error(w, "Internal error", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respJson)
	if err != nil {
		log.Printf("CreatePost() Write: %v", err)
		http.Error(w, "Internal error", http.StatusBadRequest)
	}

}

func (h *Handler) GetAllPostsUser(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	gotPosts, err := h.userRepository.GetAllPostsUser(h.userRepository.Ctx, userId, limit, offset)
	if err != nil {
		log.Printf("GetAllPostsUser() RepoGetAllPostsUser: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}

	resJson, err := json.Marshal(gotPosts)
	if err != nil {
		log.Printf("GetAllPostsUser() Marshal: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resJson)
	if err != nil {
		log.Printf("GetAllPostsUser() Write: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, _ *http.Request) {
	gotUsers, err := h.userRepository.GetAllUsers(h.userRepository.Ctx)
	if err != nil {
		log.Printf("GetAllUsers(): %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	resJson, err := json.Marshal(gotUsers)
	if err != nil {
		log.Printf("GetAllUsers() Marshal: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resJson)
	if err != nil {
		log.Printf("GetAllUsers() Write: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
