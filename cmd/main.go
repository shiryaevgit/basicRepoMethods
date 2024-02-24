package main

import (
	"fmt"
	"log"
	"myProject/database"
	"os"
)

func main() {
	dbHandler, err := database.NewHandlerDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbHandler.Close()

	fileLog, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("openFile(error.log): %v", err)
	}
	log.SetOutput(fileLog)

	err = dbHandler.SelectFromTestTable()
	if err != nil {
		log.Printf("selectFromTestTable: %v", err)
	}

	err = dbHandler.InsertIntoTestTable("Yan", 11)
	if err != nil {
		log.Printf("insertIntoTestTable: %v", err)
	}
}
