package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shiryaevgit/myProject/database"
	"github.com/shiryaevgit/myProject/pkg/models"
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

		user := new(models.User)
		err := json.NewDecoder(r.Body).Decode(user)
		if err != nil {
			log.Printf("CreateUser: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		err = h.dbHandler.Conn.QueryRow(context.Background(), "INSERT INTO users (login, full_name) VALUES ($1, $2) RETURNING *", user.Login, user.FullName).
			Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			log.Printf("CreateUser() QueryRow: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		respJson, err := json.Marshal(user)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respJson)
		if err != nil {
			log.Printf("CreateUser() Marshal(): %v", err)
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
			log.Printf("GetUserById: %v", err)
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}

		var user models.User

		err = h.dbHandler.Conn.QueryRow(context.Background(), "SELECT id, login, full_name, created_at FROM users WHERE id=$1", idInt).
			Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
		if err != nil {
			log.Printf("GetUserById:  %v", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		responseJSON, err := json.Marshal(user)
		if err != nil {
			log.Printf("GetUserById: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(responseJSON)
		if err != nil {
			log.Printf("GetUserById: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		fmt.Println(user)
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
		rows, err := h.dbHandler.Conn.Query(context.Background(), sqlQuery)
		if err != nil {
			http.Error(w, "The entered data is incorrect", http.StatusBadRequest)
			log.Printf("GetUsersList():%v", err)
			return
		}

		var users []models.User
		for rows.Next() {
			user := *new(models.User)
			err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				log.Printf("GetUsersList():%v", err)
				return
			}
			users = append(users, user)
		}

		respJson, err := json.Marshal(users)
		if err != nil {
			fmt.Printf("GetUsersList() Marshal(): %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respJson)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("GetUsersList() Write(): %v", err)
		}

		//for _, u := range users {
		//	fmt.Printf("ID:%v\nLogin:%v\nFullName:%v\nCreated at:%v\n\n", u.ID, u.Login, u.FullName, u.CreatedAt.Format("2006-01-02 15:04:05"))
		//}
	}
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		defer r.Body.Close()

		post := new(models.Post)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(post)
		if err != nil {
			http.Error(w, `Invalid request entered`, http.StatusBadRequest)
			log.Printf("CreatePost(): %v", err)
			return
		}

		var userId int
		err = h.dbHandler.Conn.QueryRow(context.Background(), "SELECT id FROM users WHERE id=$1", post.UserId).Scan(&userId)
		if err != nil {
			http.Error(w, "User not found", http.StatusBadRequest)
			log.Printf("CreatePost() QueryRow() SELECT: %v", err)
			return
		}

		err = h.dbHandler.Conn.QueryRow(context.Background(), "INSERT INTO posts (user_id,text) VALUES ($1,$2) RETURNING *", post.UserId, post.Text).
			Scan(&post.ID, &post.UserId, &post.Text, &post.CreatedAt)
		if err != nil {
			http.Error(w, "Internal error", http.StatusBadRequest)
			log.Printf("CreatePost() QueryRow() INSERT: %v", err)
			return
		}

		respJson, err := json.Marshal(post)
		if err != nil {
			http.Error(w, "Internal error", http.StatusBadRequest)
			log.Printf("CreatePost() Marshal(): %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respJson)
		if err != nil {
			http.Error(w, "Internal error", http.StatusBadRequest)
			log.Printf("CreatePost() Write(): %v", err)
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
			http.Error(w, "incorrect id", http.StatusBadRequest)
			log.Printf("GetAllPostsUser() incorrect id")
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

		rows, err := h.dbHandler.Conn.Query(context.Background(), sqlQuery)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("GetAllPostsUser() Query()%v", err)
			return
		}
		defer rows.Close()

		var posts []models.Post
		for rows.Next() {
			var post models.Post
			err = rows.Scan(&post.ID, &post.UserId, &post.Text, &post.CreatedAt)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Printf("GetAllPostsUser() Scan()%v", err)
				return
			}
			posts = append(posts, post)
		}

		resJson, err := json.Marshal(posts)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("GetAllPostsUser() Marshal(): %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(resJson)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("GetAllPostsUser() Write(): %v", err)
		}
	}
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	h.dbHandler.Mu.Lock()
	defer h.dbHandler.Mu.Unlock()

	if r.Method == http.MethodGet {
		rows, err := h.dbHandler.Conn.Query(context.Background(), "SELECT *FROM users")
		if err != nil {
			log.Printf("GetAllUsers: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var users []models.User

		for rows.Next() {
			var user models.User
			err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
			if err != nil {
				log.Printf("GetAllUsers: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			users = append(users, user)
		}

		for _, u := range users {
			fmt.Printf("ID: %v\nLogin: %v\nFullName: %v\nCreatedAt: %v\n\n", u.ID, u.Login, u.FullName, u.CreatedAt.Format("2006-01-02 15:04:05"))

		}
	}
}
