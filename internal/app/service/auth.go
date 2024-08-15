package svc

import "net/http"

func Register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("You are welcome"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello friend"))
}
