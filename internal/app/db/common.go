package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const kSchemaPrefix = "id"

// RecordType -- тип описывает записи которые мы хранить как секреты
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

// User -- тип для хранения информации о пользователе который запрашивает данные
type User struct {
	ID       string `json:"id,omitempty"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Password -- типа для хранения пароля
type Password struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Text -- тип для произвольной текстовой информации
type Text struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// Binary -- типа для хранения бинарных данных
type Binary struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Card -- тип для хранения данных о банковских картах
type Card struct {
	Name   string `json:"name"`
	Number string `json:"number"`
	Cvv    string `json:"cvv"`
	Month  string `json:"month"`
	Year   string `json:"year"`
}

// Record -- абстрактный типа с помощью которого мы передаем данные от клиента на сервер
type Record struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

var recordTypes map[string]RecordType = map[string]RecordType{
	SRecordUnknown:  RecordUnknown,
	SRecordPassword: RecordPassword,
	SRecordText:     RecordText,
	SRecordBinary:   RecordBinary,
	SRecordCard:     RecordCard,
}

// GetRecordType -- функция получает значение типа по текстовому педставлению
func GetRecordType(r string) RecordType {
	t, ok := recordTypes[r]
	if !ok {
		return RecordUnknown
	}
	return t
}

// GetSRecordType -- функция получает текстовое представление типа по значению
func GetSRecordType(rtype RecordType) string {
	for sd, t := range recordTypes {
		if t == rtype {
			return sd
		}
	}
	return SRecordUnknown
}

// PG -- структура реализующая интерфейс Storage.
type PG struct {
	DB     *pgxpool.Pool
	Logger *log.Logger
	Secret string
}

// Storage -- интерфейс харнения секретов
type Storage interface {
	Register(ctx context.Context, u *User) error
	Login(ctx context.Context, u *User) error
	List(ctx context.Context, u *User) ([]Record, error)
	Put(ctx context.Context, u *User, rec *Record) error
	Get(ctx context.Context, u *User, rec *Record) (*Record, error)
}
