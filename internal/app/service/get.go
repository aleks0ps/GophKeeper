package svc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	mycookie "github.com/aleks0ps/GophKeeper/internal/app/cookie"
	"github.com/aleks0ps/GophKeeper/internal/app/db"

	myhttp "github.com/aleks0ps/GophKeeper/internal/app/http"
)

func (s *Svc) Get(w http.ResponseWriter, r *http.Request) {
	err := mycookie.ValidateCookie(r)
	if err != nil {
		myhttp.WriteError(w, http.StatusUnauthorized, err)
	}
	userID, _ := mycookie.GetCookie(r, "id")
	defer r.Body.Close()
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(r.Body)
	if err != nil && err != io.EOF {
		s.Logger.Println("ERR:svc:Get: ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	var rec db.Record
	err = json.Unmarshal(buf.Bytes(), &rec)
	if err != nil {
		s.Logger.Println("ERR:svc:Get ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	var res *db.Record
	user := db.User{ID: userID}
	res, err = s.DB.Get(r.Context(), &user, &rec)
	if err != nil {
		s.Logger.Println("ERR:svc:Get ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	s.Logger.Printf("%v\n", *res)
	myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusOK, nil)
}
