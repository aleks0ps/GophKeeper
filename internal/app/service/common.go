package svc

import (
	"log"

	"github.com/aleks0ps/GophKeeper/internal/app/db"
)

// Svc -- структура описывает основные параметры сервиса
type Svc struct {
	Logger  *log.Logger
	DB      db.Storage
	DataDir string
}
