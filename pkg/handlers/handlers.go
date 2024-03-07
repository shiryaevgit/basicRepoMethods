package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/doug-martin/goqu/v9"
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

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("CreateUser(): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	sqlQuery := "INSERT INTO users (login, full_name) VALUES ($1, $2) RETURNING *"

	createdUser, err := h.dbHandler.RepoInsertUser(h.dbHandler.Ctx, user, sqlQuery)
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

func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {

	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	//idString := r.URL.Path[len("/users/"):]
	idString := r.PathValue("id")
	fmt.Println(idString)
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		log.Printf("GetUserById(): %v", err)
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	sqlQuery := "SELECT id, login, full_name, created_at FROM users WHERE id=$1"

	gotUser, err := h.dbHandler.RepoGetUserById(h.dbHandler.Ctx, idInt, sqlQuery)
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

func (h *Handler) GetUsersList(w http.ResponseWriter, r *http.Request) {
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	orderBy := r.URL.Query().Get("orderBy")
	login := r.URL.Query().Get("login")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

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
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
		}
		ds = ds.Limit(uint(limitInt))
	}
	if offset != "" {
		offsetInt, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			http.Error(w, "invalid offset parameter", http.StatusBadRequest)
		}
		ds = ds.Offset(uint(offsetInt))
	}

	sqlQuery, _, err := ds.ToSQL()
	if err != nil {
		log.Printf("GetUsersList() ToSQL: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
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

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	post := new(models.Post)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(post)
	if err != nil {
		log.Printf("CreatePost(): %v", err)
		http.Error(w, `Invalid request entered`, http.StatusBadRequest)
		return
	}
	sqlQueryCheck := "SELECT id FROM users WHERE id=$1"
	if err = h.dbHandler.RepoCheckUser(h.dbHandler.Ctx, post.UserId, sqlQueryCheck); err != nil {
		log.Printf("CreatePost() RepoCheckUser: %v", err)
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	sqlQueryCreated := "INSERT INTO posts (user_id,text) VALUES ($1,$2) RETURNING *"
	createdPost, err := h.dbHandler.RepoCreatePost(h.dbHandler.Ctx, *post, sqlQueryCreated)
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
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	userId := r.URL.Query().Get("userId")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	ds := goqu.From("posts")

	if userId != "" {
		ds = ds.Where(goqu.C("user_id").Eq(userId))
	}
	if limit != "" {
		limitInt, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
		}
		ds = ds.Limit(uint(limitInt))
	}

	if offset != "" {
		offsetInt, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			http.Error(w, "invalid offset parameter", http.StatusBadRequest)
		}
		ds = ds.Offset(uint(offsetInt))
	}
	sqlQuery, _, _ := ds.ToSQL()

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

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	sqlQuery := "SELECT *FROM users"
	gotUsers, err := h.dbHandler.RepoGetAllUsers(h.dbHandler.Ctx, sqlQuery)
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
