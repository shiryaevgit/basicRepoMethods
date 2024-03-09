package main

import (
	"context"
	"errors"
	"github.com/shiryaevgit/basicRepoMethods/config"
	"github.com/shiryaevgit/basicRepoMethods/pkg/handlers"
	"github.com/shiryaevgit/basicRepoMethods/pkg/loggers/logrus"
	"github.com/shiryaevgit/basicRepoMethods/pkg/loggers/standLog"
	"github.com/shiryaevgit/basicRepoMethods/pkg/server"
	"github.com/shiryaevgit/basicRepoMethods/repository/postgres"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

func main() {

	terminateContext, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFunc()

	configFile, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("LoadConfig(): %v", err)
	}

	db, err := postgres.NewRepoPostgres(terminateContext, configFile.PostgresURL)
	if err != nil {
		log.Fatalf("unable to connect to postgres: %v", err)
	}
	defer db.Close()

	// standLog
	fileLog, err := standLog.LoadStandLog("standLog.log")
	if err != nil {
		log.Fatalf("LoadStandLog():%v", err)
	}
	defer fileLog.Close()

	// logrus
	logger, fileLogrus, err := logrus.SetupLogger("logrus.log")
	if err != nil {
		log.Fatalf("SetupLogger: %v", err)
	}
	logger.Info("This log message is configured using loggerconfig package")
	defer fileLogrus.Close()

	srv := new(server.Server)
	mux := http.NewServeMux()
	handlerDb := handlers.NewHandlerServ(db)

	mux.HandleFunc("POST /users", handlerDb.CreateUser)
	mux.HandleFunc("GET /users/{id}", handlerDb.GetUserById)
	mux.HandleFunc("GET /users/all", handlerDb.GetAllUsers)
	mux.HandleFunc("GET /users", handlerDb.GetUsersList)
	mux.HandleFunc("POST /posts", handlerDb.CreatePost)
	mux.HandleFunc("GET /posts", handlerDb.GetAllPostsUser)

	portStr := strconv.Itoa(configFile.HTTPPort)
	err = srv.Run(portStr, mux, terminateContext)
	switch {
	case err != nil && errors.Is(err, http.ErrServerClosed):
		log.Printf("Run(): %v", err)
		logger.Error("Run(): ", err)
		logger.Info("Run(): ", err)

	case err != nil:
		log.Printf("Run(): %v", err)
	default:
		log.Printf("Server is running on http://127.0.0.1%v\n", configFile.HTTPPort)
	}
}
