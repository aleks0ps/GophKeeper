package svc

import (
	"encoding/json"
	"net/http"

	mycookie "github.com/aleks0ps/GophKeeper/internal/app/cookie"
	"github.com/aleks0ps/GophKeeper/internal/app/db"

	myhttp "github.com/aleks0ps/GophKeeper/internal/app/http"
)

// List -- функция возвращает все секреты которые есть у пользователя не раскрывая самих данных
func (s *Svc) List(w http.ResponseWriter, r *http.Request) {
	err := mycookie.ValidateCookie(r)
	if err != nil {
		myhttp.WriteError(w, http.StatusUnauthorized, err)
	}
	userID, _ := mycookie.GetCookie(r, "id")
	recs, err := s.DB.List(r.Context(), &db.User{ID: userID})
	if err != nil {
		s.Logger.Println("ERR:svc:List: ", err)
		myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusInternalServerError, nil)
		return
	}
	payload, err := json.Marshal(recs)
	if err != nil {
		s.Logger.Println("ERR:svc:List: ", err)
		myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusInternalServerError, nil)
		return
	}
	myhttp.WriteResponse(w, myhttp.CTypeJSON, http.StatusOK, payload)
}
