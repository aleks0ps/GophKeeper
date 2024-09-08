package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const kSchemaPrefix = "id"

type User struct {
	ID       string `json:"id,omitempty"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Record struct {
	Data []byte `json:"data"`
}

// Storage
type PG struct {
	DB     *pgxpool.Pool
	Logger *log.Logger
}

type Storage interface {
	Register(ctx context.Context, u *User) error
	Login(ctx context.Context, u *User) error
	//Store(ctx context.Context, u *User, rec *Record) error
	//Load(ctx context.Context, u *User, rec *Record) error
	//Update(ctx context.Context, u *User, rec *Record) error
	//Delete(ctx context.Context, u *User, rec *Record) error
}
