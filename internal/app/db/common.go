package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const kSchemaPrefix = "id"

type RecordType int

const (
	RecordUnknown RecordType = iota
	RecordPassword
	RecordCard
	RecordText
	RecordBinary
)

const (
	SRecordUnknown  = "unknown"
	SRecordPassword = "password"
	SRecordCard     = "card"
	SRecordText     = "text"
	SRecordBinary   = "binary"
)

type User struct {
	ID       string `json:"id,omitempty"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Password struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Binary struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Record struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

// Storage
type PG struct {
	DB     *pgxpool.Pool
	Logger *log.Logger
}

type Storage interface {
	Register(ctx context.Context, u *User) error
	Login(ctx context.Context, u *User) error
	List(ctx context.Context, u *User) ([]Record, error)
	Put(ctx context.Context, u *User, rec *Record) error
	//Load(ctx context.Context, u *User, rec *Record) error
	//Update(ctx context.Context, u *User, rec *Record) error
	//Delete(ctx context.Context, u *User, rec *Record) error
}
