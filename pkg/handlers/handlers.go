package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/shiryaevgit/basicRepoMethods/database"
	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	dbHandler *database.UserRepository
}

func NewHandlerServ(db *database.UserRepository) *Handler {
	return &Handler{dbHandler: db}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Printf("CreateUser(): %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		createdUser, err := h.dbHandler.RepoInsertUser(h.dbHandler.Ctx, user)
		if err != nil {
			log.Printf("CreateUser(): %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		respJson, err := json.Marshal(createdUser)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respJson)
		if err != nil {
			log.Printf("CreateUser() Marshal: %v", err)
		}
	}
}

func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	if r.Method == http.MethodGet {

		id := r.URL.Path[len("/users/"):]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("GetUserById(): %v", err)
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}

		gotUser, err := h.dbHandler.RepoGetUserById(h.dbHandler.Ctx, idInt)
		if err != nil {
			log.Printf("GetUserById(): %v", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		responseJSON, err := json.Marshal(gotUser)
		if err != nil {
			log.Printf("GetUserById(): %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(responseJSON)
		if err != nil {
			log.Printf("GetUserById(): %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) GetUsersList(w http.ResponseWriter, r *http.Request) {
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	if r.Method == http.MethodGet {
		orderBy := r.URL.Query().Get("orderBy")
		login := r.URL.Query().Get("login")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		sqlQuery := "SELECT * FROM users"
		if login != "" {
			sqlQuery += fmt.Sprintf(" WHERE login='%s'", login)

		}
		if orderBy != "" {
			sqlQuery += fmt.Sprintf(" ORDER BY %s", orderBy)

		}
		if limit != "" {
			sqlQuery += fmt.Sprintf(" LIMIT %s", limit)
		}
		if offset != "" {
			sqlQuery += fmt.Sprintf(" OFFSET %s", offset)
		}

		gotUsers, err := h.dbHandler.RepoGetUsersList(h.dbHandler.Ctx, sqlQuery)

		if err != nil {
			log.Printf("GetUsersList() RepoGetUsersList: %v", err)
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
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		defer r.Body.Close()

		post := new(models.Post)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(post)
		if err != nil {
			log.Printf("CreatePost(): %v", err)
			http.Error(w, `Invalid request entered`, http.StatusBadRequest)
			return
		}

		createdPost, err := h.dbHandler.RepoCreatePost(h.dbHandler.Ctx, *post)
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
}

func (h *Handler) GetAllPostsUser(w http.ResponseWriter, r *http.Request) {
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	if r.Method == http.MethodGet {
		userId := r.URL.Query().Get("userId")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		sqlQuery := fmt.Sprintf("SELECT * FROM posts")
		if userId == "" {
			log.Printf("GetAllPostsUser() incorrect id")
			http.Error(w, "incorrect id", http.StatusBadRequest)
			return
		} else {
			sqlQuery += fmt.Sprintf(" WHERE user_id='%s'", userId)
		}
		if limit != "" {
			sqlQuery += fmt.Sprintf(" LIMIT %s", limit)
		}
		if offset != "" {
			sqlQuery += fmt.Sprintf(" OFFSET %s", offset)
		}

		gotPosts, err := h.dbHandler.RepoGetAllPostsUser(h.dbHandler.Ctx, sqlQuery)
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
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	if r.Method == http.MethodGet {

		gotUsers, err := h.dbHandler.RepoGetAllUsers(h.dbHandler.Ctx)
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
}
