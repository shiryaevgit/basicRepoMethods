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
		tempUser := new(models.User)
		err := json.NewDecoder(r.Body).Decode(tempUser)
		if err != nil {
			log.Printf("CreateUser: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		_, err = h.dbHandler.Сonn.Exec(context.Background(), "INSERT INTO users (login, full_name) VALUES ($1, $2)", tempUser.Login, tempUser.FullName)
		if err != nil {
			log.Printf("CreateUser: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		id := r.URL.Path[len("/users/"):]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("GetUserById: %v", err)
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}

		var user models.User

		err = h.dbHandler.Сonn.QueryRow(context.Background(), "SELECT id, login, full_name, created_at FROM users WHERE id=$1", idInt).
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
		rows, err := h.dbHandler.Сonn.Query(context.Background(), sqlQuery)
		if err != nil {
			http.Error(w, "The entered data is incorrect", http.StatusBadRequest)
			log.Printf("GetUsersList():%v", err)
		}

		var users []models.User
		for rows.Next() {
			user := *new(models.User)
			err = rows.Scan(&user.ID, &user.Login, &user.FullName, &user.CreatedAt)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				log.Printf("GetUsersList():%v", err)
			}
			users = append(users, user)
		}

		for _, u := range users {
			fmt.Printf("ID:%v\nLogin:%v\nFullName:%v\nCreated at:%v\n\n", u.ID, u.Login, u.FullName, u.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
}

/*
	3. Получение списка пользователей:
	GET /users?orderBy=...&login=...&limit=...&offset=...
	Query параметры в запросе:
	orderBy - сортировка запрашиваемых данных по колонкам: CreatedAt, Login.
	login - логин пользователя к выдаче
	limit - кол-во пользователей к выдаче в запросе
	offset - кол-во пользователей к пропуску при выдаче

	Все query параметры опциональны (могут как быть переданы в запросе, так и опущены).

	* при реализации limit, offset советую изучить что такое Пагинация.
*/

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreatePost")
}

func (h *Handler) GetAllPostsUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllPostsUser")

}
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, err := h.dbHandler.Сonn.Query(context.Background(), "SELECT *FROM users")
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
