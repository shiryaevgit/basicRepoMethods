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
	fmt.Println(configFile.DatabaseURL)
	fmt.Println(configFile.HTTPPort)

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

	err = dbHandler.SelectFromTestTable()
	if err != nil {
		log.Printf("selectFromTestTable: %v", err)
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
