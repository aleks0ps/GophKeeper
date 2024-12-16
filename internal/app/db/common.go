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

type Text struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type Binary struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Card struct {
	Name   string `json:"name"`
	Number string `json:"number"`
	Cvv    string `json:"cvv"`
	Month  string `json:"month"`
	Year   string `json:"year"`
}

type Record struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

// alias for Record
type Data Record

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
	Get(ctx context.Context, u *User, rec *Record) (*Record, error)
}
