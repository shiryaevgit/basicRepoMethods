package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/shiryaevgit/basicRepoMethods/config"
	"github.com/shiryaevgit/basicRepoMethods/pkg/handlers"
	"github.com/shiryaevgit/basicRepoMethods/pkg/loggers/logrus"
	"github.com/shiryaevgit/basicRepoMethods/pkg/loggers/standLog"
	"github.com/shiryaevgit/basicRepoMethods/pkg/server"
	"github.com/shiryaevgit/basicRepoMethods/repository/mongo"
	"github.com/shiryaevgit/basicRepoMethods/repository/postgres"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFunc()

	configFile, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("LoadConfig(): %v", err)
	}

	// Пример того как можно динамически выбирать монгу или пг
	// Заместо true - реализуй свое условие (например, читая из конфига)
	var userRepository handlers.UserRepository
	if true {
		dataBasePostgres, err := postgres.NewRepoPostgres(ctx, configFile.PostgresURL)
		if err != nil {
			log.Fatalf("unable to connect to postgres: %v", err)
		}
		defer dataBasePostgres.Close()

		userRepository = dataBasePostgres
	} else {
		dataBaseMongo, err := mongo.NewRepoMongo(ctx, configFile.MongoURI)
		if err != nil {
			log.Fatalf("unable to connect to mongo: %v", err)
		}
		defer dataBaseMongo.Close()

		userRepository = dataBaseMongo
	}

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
	handlerDb := handlers.NewHandlerServ(userRepository)

	mux.HandleFunc("POST /users", handlerDb.CreateUser)
	mux.HandleFunc("GET /users/{id}", handlerDb.GetUserById)
	mux.HandleFunc("GET /users/all", handlerDb.GetAllUsers)
	mux.HandleFunc("GET /users", handlerDb.GetUsersList)
	mux.HandleFunc("POST /posts", handlerDb.CreatePost)
	mux.HandleFunc("GET /posts", handlerDb.GetAllPostsUser)

	portStr := strconv.Itoa(configFile.HTTPPort)
	err = srv.Run(portStr, mux, ctx)
	switch {
	case err != nil && errors.Is(err, http.ErrServerClosed):
		log.Printf("Run(): %v", err)
		logger.Error("Run(): ", err)
		logger.Info("Run(): ", err)
	case err != nil:
		log.Printf("Run(): %v", err)
	default:
		log.Printf("Server is running on http://127.0.0.1:%d", configFile.HTTPPort)
	}
}
