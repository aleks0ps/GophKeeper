package svc

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"

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
	var resp *db.Record
	user := db.User{ID: userID}
	resp, err = s.DB.Get(r.Context(), &user, &rec)
	if err != nil {
		s.Logger.Println("ERR:svc:Get ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if resp.Type == db.SRecordBinary {
		var binary db.Binary
		// Decode response payload
		err = json.Unmarshal(resp.Payload, &binary)
		if err != nil {
			s.Logger.Println("ERR:svc:Get: ", err)
			myhttp.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		writer := multipart.NewWriter(w)
		contentType := writer.FormDataContentType()
		w.Header().Set("Content-Type", contentType)
		file, err := os.Open(binary.Path)
		if err != nil {
			s.Logger.Println("ERR:svc:Get: ", err)
			return
		}
		defer file.Close()
		defer writer.Close()
		part, err := writer.CreateFormFile("file", binary.Name)
		if err != nil {
			s.Logger.Println("ERR:svc:Get: ", err)
			return
		}
		for {
			_, err = io.CopyN(part, file, 4096)
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				break
			}
			if err != nil {
				s.Logger.Println("ERR:svc:Get: ", err)
				return
			}
		}
		s.Logger.Println("INFO:svc:Get: file uploaded to client")
	} else {
		payload, err := json.Marshal(resp)
		if err != nil {
			s.Logger.Println("ERR:svc:Get: ", err)
			myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusInternalServerError, nil)
			return
		}
		myhttp.WriteResponse(w, myhttp.CTypeJSON, http.StatusOK, payload)
	}
}
