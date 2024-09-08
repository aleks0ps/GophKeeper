package svc

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/aleks0ps/GophKeeper/internal/app/db"
	myhttp "github.com/aleks0ps/GophKeeper/internal/app/http"
)

// Debug function
func DEBUG(logger *log.Logger, r *http.Request) {
	logger.Printf("%s\n", r.Method)
	logger.Printf("%s\n", r.URL)
	for k, v := range r.Header {
		logger.Printf("%s\n", k)
		for i := 0; i < len(v); i++ {
			logger.Printf("[%s] %s\n", k, v[i])
		}
	}
}

func (s *Svc) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf := bytes.Buffer{}
	// read from Body and append it to a buffer
	n, err := buf.ReadFrom(r.Body)
	if err != nil && err != io.EOF {
		s.Logger.Println("ERR:Register: %d %v\n", n, err)
		return
	}
	// wait for json
	if r.Header.Get(myhttp.SContentType) != myhttp.STypeJSON {
		s.Logger.Println("ERR:Register: ", r.Header.Get(myhttp.SContentType))
		myhttp.WriteError(w, http.StatusBadRequest, nil)
		return
	}
	var user *db.User
	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		s.Logger.Println("ERR:Register: ", err)
		return
	}
	if err = s.DB.Register(r.Context(), user); err != nil {
		s.Logger.Println("ERR:Register: ", err)
		return
	}
}

func (s *Svc) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf := bytes.Buffer{}
	n, err := buf.ReadFrom(r.Body)
	if err != nil && err != io.EOF {
		s.Logger.Println("ERR:Login: %d %v\n", n, err)
		return
	}
	if r.Header.Get(myhttp.SContentType) != myhttp.STypeJSON {
		s.Logger.Println("'ERR:Login: ", r.Header.Get(myhttp.SContentType))
		myhttp.WriteError(w, http.StatusBadRequest, nil)
		return
	}
	var user *db.User
	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		s.Logger.Println("ERR:Login: ", err)
		return
	}
	if err = s.DB.Login(r.Context(), user); err != nil {
		s.Logger.Println("ERR:Login: ", err)
		return
	}
}

func (s *Svc) Health(w http.ResponseWriter, r *http.Request) {
	s.Logger.Printf("OK\n")
	w.Write([]byte("OK"))
}
