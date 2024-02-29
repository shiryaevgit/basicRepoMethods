package main

import (
	"context"
	"errors"
	"github.com/shiryaevgit/myProject/config"
	"github.com/shiryaevgit/myProject/database"
	"github.com/shiryaevgit/myProject/pkg/handlers"
	"github.com/shiryaevgit/myProject/pkg/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

func main() {

	configFile, err := config.LoadConfig("conf.json")
	if err != nil {
		log.Fatalf("config.LoadConfig(): %v", err)
	}

	db, err := database.NewHandlerDB(configFile.DatabaseURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer db.Close()

	fileLog, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("openFile(error.log): %v", err)
	}
	defer fileLog.Close()

	log.SetOutput(fileLog)

	srv := new(server.Server)
	mux := http.NewServeMux()
	handlerDb := handlers.NewHandlerServ(db)

	mux.HandleFunc("/users/all", handlerDb.GetAllUsers)
	mux.HandleFunc("/users/", handlerDb.GetUserById)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlerDb.GetUsersList(w, r)
		} else if r.Method == http.MethodPost {
			handlerDb.CreateUser(w, r)
		}
	})
	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlerDb.CreatePost(w, r)
		} else if r.Method == http.MethodGet {
			handlerDb.GetAllPostsUser(w, r)
		}
	})

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	portStr := strconv.Itoa(configFile.HTTPPort)
	err = srv.Run(portStr, mux, ctx)
	switch {
	case err != nil && errors.Is(err, http.ErrServerClosed):
		log.Printf("Run(): %v", err)
	case err != nil:
		log.Printf("Run(): %v", err)
	default:
		log.Printf("Server is running on http://127.0.0.1%v\n", configFile.HTTPPort)
	}

}

/*
Требования к моделям данных:
Пользователь:
ID, CreatedAt(дата создания), Login, ФИО
Пост
ID, CreatedAt, UserID(пользователь, создавший пост), Text(текст поста)

Все поля NOT NULL
Поля ID, CreatedAt должны генерироваться на стороне БД. Пример:
https://www.postgresqltutorial.com/postgresql-tutorial/postgresql-identity-column/

Возвращать сгенерированные БД значения можно используя оператор RETURNING:

INSERT INTO ... RETURNING *;


Методы HTTP сервера, обеспечивающие взаимодействие с БД:

1. Создание пользователя:
POST /users
Принимает в теле поля, требуемые для создания пользователя, возвращает созданного пользователя.

2. Получение пользователя:
GET /users/{id}
В качестве PATH параметра принимает идентификатор пользователя.
Возвращает в ответе созданного пользователя.

3. Получение списка пользователей:
GET /users?orderBy=...&login=...&limit=...&offset=...
Query параметры в запросе:
orderBy - сортировка запрашиваемых данных по колонкам: CreatedAt, Login
login - логин пользователя к выдаче
limit - кол-во пользователей к выдаче в запросе
offset - кол-во пользователей к пропуску при выдаче

Все query параметры опциональны (могут как быть переданы в запросе, так и опущены).

* при реализации limit, offset советую изучить что такое Пагинация.

4. Создание поста пользователем:
POST /posts
Принимает в теле поля, требуемые для создания поста, возвращает созданный пост.

5. Получение списка постов пользователя:
GET /posts?userId=...&limit=...&offset=...
Возвращает созданные пользователями посты.

Query параметры в запросе:
userId - фильтр на идентификатор пользователя, создавшего посты
limit - кол-во постов к выдаче в запросе
offset - кол-во постов к пропуску при выдаче

Все query параметры опциональны (могут как быть переданы в запросе, так и опущены).

CREATE TABLE public.users (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL,
    full_name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.posts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES public.users(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);




*/
