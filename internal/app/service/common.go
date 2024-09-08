package svc

import (
	"log"

	"github.com/aleks0ps/GophKeeper/internal/app/db"
)

type Svc struct {
	Logger *log.Logger
	DB     db.Storage
}
