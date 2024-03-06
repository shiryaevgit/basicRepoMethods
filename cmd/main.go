package main

import (
	"context"
	"errors"
	"github.com/shiryaevgit/basicRepoMethods/config"
	"github.com/shiryaevgit/basicRepoMethods/database"
	"github.com/shiryaevgit/basicRepoMethods/pkg/handlers"
	"github.com/shiryaevgit/basicRepoMethods/pkg/loggers/logrus"
	"github.com/shiryaevgit/basicRepoMethods/pkg/loggers/standLog"
	"github.com/shiryaevgit/basicRepoMethods/pkg/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	terminateContext, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFunc() // для чего здесь ?

	configFile, err := config.LoadConfig("conf.json")
	if err != nil {
		log.Fatalf("config.LoadConfig(): %v", err)
	}

	db, err := database.NewUserRepository(terminateContext, configFile.DatabaseURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
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
