package main

import (
	"fmt"
	"github.com/shiryaevgit/myProject/config"
	"github.com/shiryaevgit/myProject/database"
	"log"
	"os"
)

func main() {

	configFile, err := config.LoadConfig("conf.json")
	if err != nil {
		log.Fatalf("config.LoadConfig(): %v", err)
	}

	dbHandler, err := database.NewHandlerDB(configFile.DatabaseURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer dbHandler.Close()

	fileLog, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("openFile(error.log): %v", err)
	}
	log.SetOutput(fileLog)

	users, err := dbHandler.GetAllUsers()
	if err != nil {
		log.Printf("GetAllUsers: %v", err)
	}
	for _, user := range users {
		fmt.Printf("ID: %v\nLogin: %v\nFullName: %v\nCreated: %s\n\n", user.ID, user.Login, user.FullName, user.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	posts, err := dbHandler.GetAllPosts(1)
	if err != nil {
		log.Printf("GetAllPosts: %v", err)
	}
	for _, post := range posts {
		fmt.Printf("ID: %v\n UserId: %v\n Text: %v\n Created: %s\n", post.ID, post.UserId, post.Text, post.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	err = dbHandler.InsertIntoTestTable("yan", "Yanush Chernyih")
	if err != nil {
		log.Printf("insertIntoTestTable: %v", err)
	}
}

/*
mydb=# CREATE TABLE public.users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    login TEXT NOT NULL,
    full_name TEXT NOT NULL
);
CREATE TABLE
mydb=# CREATE TABLE public.posts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES public.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    text TEXT NOT NULL
);

.env
DATABASE_URL=postgres://user:zaq1xsw2@localhost:5432/mydb?sslmode=disable


*/
