package svc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	mycookie "github.com/aleks0ps/GophKeeper/internal/app/cookie"
	"github.com/aleks0ps/GophKeeper/internal/app/db"
	"github.com/docker/go-units"

	myhttp "github.com/aleks0ps/GophKeeper/internal/app/http"
)

// Put -- записывает секрет пользователя в хранилище
func (s *Svc) Put(w http.ResponseWriter, r *http.Request) {
	err := mycookie.ValidateCookie(r)
	if err != nil {
		myhttp.WriteError(w, http.StatusUnauthorized, err)
	}
	userID, _ := mycookie.GetCookie(r, "id")
	defer r.Body.Close()
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(r.Body)
	if err != nil && err != io.EOF {
		s.Logger.Println("ERR:svc:Put: ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	var rec db.Record
	err = json.Unmarshal(buf.Bytes(), &rec)
	if err != nil {
		s.Logger.Println("ERR:svc:Put ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	user := db.User{ID: userID}
	err = s.DB.Put(r.Context(), &user, &rec)
	if err != nil {
		s.Logger.Println("ERR:svc:Put ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusOK, nil)
}

// PutBinary -- функция предназначения для загрузки бинарных данные произвольного размера, в хранилище записывается путь к данным на диске
func (s *Svc) PutBinary(w http.ResponseWriter, r *http.Request) {
	err := mycookie.ValidateCookie(r)
	if err != nil {
		myhttp.WriteError(w, http.StatusUnauthorized, err)
	}
	userID, _ := mycookie.GetCookie(r, "id")
	defer r.Body.Close()
	form, h, err := r.FormFile("file")
	if err != nil {
		s.Logger.Println("ERR:svc:PutBinary: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// make dir for uploads
	userDir := s.DataDir + "/" + userID
	err = os.Mkdir(userDir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		s.Logger.Println("ERR:svc:PutBinary")
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	// encode filename to base64
	encoded := base64.StdEncoding.EncodeToString([]byte(h.Filename))
	pr, pw := io.Pipe()
	filePath := userDir + "/" + encoded
	file, err := os.Create(filePath)
	if err != nil {
		s.Logger.Println("ERR:svc:PutBinary ", err)
		return
	}
	go func() {
		defer pw.Close()
		// 4 Kib buffer
		b := make([]byte, 4*units.KiB)
		buf := bytes.NewBuffer(b)
		for {
			// read from http socket
			if _, err = io.ReadAtLeast(form, buf.Bytes(), 1); err != nil {
				// no bytes were read
				if err == io.ErrUnexpectedEOF || err == io.EOF {
					break
				}
				s.Logger.Println("ERR:svc:PutBinary ", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// write to a pipe
			if _, err = pw.Write(buf.Bytes()); err != nil {
				s.Logger.Println("ERR:svc:PutBinary ", err)
				break
			}
		}
	}()
	b := make([]byte, 4*units.KiB)
	buf := bytes.NewBuffer(b)
	for {
		if _, err = pr.Read(buf.Bytes()); err != nil {
			if err == io.EOF {
				break
			}
			s.Logger.Println("ERR:svc:PutBinary ", err)
			return
		}
		_, err := file.Write(buf.Bytes())
		if err != nil {
			s.Logger.Println("ERR:svc:PutBinary ", err)
			return
		}
	}
	var rec db.Record
	var binary db.Binary
	binary.Name = h.Filename
	binary.Path = filePath
	payload, err := json.Marshal(&binary)
	if err != nil {
		s.Logger.Println("ERR:svc:PutBinary ", err)
		return
	}
	rec.Type = db.SRecordBinary
	rec.Payload = payload
	user := db.User{ID: userID}
	err = s.DB.Put(r.Context(), &user, &rec)
	if err != nil {
		s.Logger.Println("ERR:svc:PutBinary ", err)
		myhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	myhttp.WriteResponse(w, myhttp.CTypeNone, http.StatusOK, nil)
}
