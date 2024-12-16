package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/aleks0ps/GophKeeper/cmd/gophkeeper/config"
	"github.com/aleks0ps/GophKeeper/internal/app/db"
	svc "github.com/aleks0ps/GophKeeper/internal/app/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opts := config.ParseOptions()
	// set up logger
	logger := log.New(os.Stdout, "SVC: ", log.LstdFlags)
	// set up storage
	db, err := db.NewDB(ctx, opts.DatabaseURI, logger)
	if err != nil {
		logger.Fatal(err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	dataDir := pwd + "/data"
	err = os.Mkdir(dataDir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		logger.Fatal(err)
	}
	service := svc.Svc{Logger: logger, DB: db, DataDir: dataDir}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/register", service.Register)
	r.Post("/login", service.Login)
	r.Post("/list", service.List)
	r.Post("/put", service.Put)
	r.Post("/put/binary", service.PutBinary)
	http.ListenAndServe(":8080", r)
}
