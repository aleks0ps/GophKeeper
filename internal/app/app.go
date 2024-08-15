package app

import (
	"net/http"

	svc "github.com/aleks0ps/GophKeeper/internal/app/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Run() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/register", svc.Register)
	r.Post("/login", svc.Login)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	http.ListenAndServe(":8080", r)
}
