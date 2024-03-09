package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/shiryaevgit/basicRepoMethods/pkg/models"
	"github.com/shiryaevgit/basicRepoMethods/repository/postgres"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	dbHandler *postgres.RepoPostgres
}

func NewHandlerServ(db *postgres.RepoPostgres) *Handler {
	return &Handler{dbHandler: db}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("CreateUser(): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	createdUser, err := h.dbHandler.CreateUser(h.dbHandler.Ctx, user)
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

	idString := r.PathValue("id")
	fmt.Println(idString)
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		log.Printf("GetUserById(): %v", err)
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	gotUser, err := h.dbHandler.GetUserById(h.dbHandler.Ctx, idInt)
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

	login := r.URL.Query().Get("login")
	orderBy := r.URL.Query().Get("orderBy")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	gotUsers, err := h.dbHandler.GetUsersList(h.dbHandler.Ctx, login, orderBy, limit, offset)

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

	if err = h.dbHandler.CheckUser(h.dbHandler.Ctx, post.UserId); err != nil {
		log.Printf("CreatePost(): %v", err)
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	createdPost, err := h.dbHandler.CreatePost(h.dbHandler.Ctx, *post)
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

	gotPosts, err := h.dbHandler.GetAllPostsUser(h.dbHandler.Ctx, userId, limit, offset)
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

	gotUsers, err := h.dbHandler.GetAllUsers(h.dbHandler.Ctx)
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
