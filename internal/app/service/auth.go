// Package svc -- пакет реализует веб сервис хранения паролей
package svc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	mycookie "github.com/aleks0ps/GophKeeper/internal/app/cookie"
	"github.com/aleks0ps/GophKeeper/internal/app/db"
	myerror "github.com/aleks0ps/GophKeeper/internal/app/error"
	myhttp "github.com/aleks0ps/GophKeeper/internal/app/http"
)

// Register -- регистрирует пользователя и отдает cookies
func (s *Svc) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf := bytes.Buffer{}
	// read from Body and append it to a buffer
	n, err := buf.ReadFrom(r.Body)
	if err != nil && err != io.EOF {
		s.Logger.Printf("ERR:svc:Register: %d %v\n", n, err)
		return
	}
	// wait for json
	if r.Header.Get(myhttp.SContentType) != myhttp.STypeJSON {
		s.Logger.Println("ERR:svc:Register: ", r.Header.Get(myhttp.SContentType))
		myhttp.WriteError(w, http.StatusBadRequest, nil)
		return
	}
	var user *db.User
	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		s.Logger.Println("ERR:svc:Register: ", err)
		return
	}
	if err = s.DB.Register(r.Context(), user); err != nil {
		s.Logger.Println("ERR:svc:Register: ", err)
		return
	}
	// issue cookies
	_, err = mycookie.EnsureCookie(w, r, user.Login)
	if err != nil {
		s.Logger.Println(err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusOK, nil)
}

// Login -- логирует пользователя и отдает cookies
func (s *Svc) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf := bytes.Buffer{}
	n, err := buf.ReadFrom(r.Body)
	if err != nil && err != io.EOF {
		s.Logger.Printf("ERR:svc:Login: %d %v\n", n, err)
		return
	}
	if r.Header.Get(myhttp.SContentType) != myhttp.STypeJSON {
		s.Logger.Println("'ERR:svc:Login: ", r.Header.Get(myhttp.SContentType))
		myhttp.WriteError(w, http.StatusBadRequest, nil)
		return
	}
	var user *db.User
	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		s.Logger.Println("ERR:svc:Login: ", err)
		return
	}
	if err = s.DB.Login(r.Context(), user); err != nil {
		s.Logger.Println("ERR:svc:Login: ", err)
		if errors.Is(err, myerror.ErrInvalidLoginOrPassword) {
			myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusUnauthorized, nil)
		}
		myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusInternalServerError, nil)
		return
	}
	// issue cookie
	_, err = mycookie.EnsureCookie(w, r, user.Login)
	if err != nil {
		s.Logger.Println("ERR:svc:Login: ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusOK, nil)
}
