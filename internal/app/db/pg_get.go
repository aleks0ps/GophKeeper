package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aleks0ps/GophKeeper/internal/app/enc"
	"github.com/jackc/pgtype"
)

func (p *PG) getPassword(ctx context.Context, u *User, rec *Record) (*Record, error) {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Printf("ERR:db:getPassword: %+v\n", err)
		return nil, err
	}
	var pass Password
	resp := &Record{Type: SRecordPassword}
	err = json.Unmarshal(rec.Payload, &pass)
	if err != nil {
		p.Logger.Printf("ERR:db:getPassword: %+v\n", err)
		return nil, err
	}
	var encrypted pgtype.Bytea
	table := fmt.Sprintf("id%s.password", userID)
	sql := fmt.Sprintf(`select password from %s where name=$1`, table)
	err = p.DB.QueryRow(ctx, sql, pass.Name).Scan(&encrypted)
	if err != nil {
		p.Logger.Printf("ERR:db:getPassword: %+v\n", err)
		return nil, err
	}
	// Decrypt data
	if err != nil {
		p.Logger.Printf("ERR:db:getPassword: %+v\n", err)
		return nil, err
	}
	password, err := enc.Decrypt([]byte(p.Secret), encrypted.Bytes)
	if err != nil {
		p.Logger.Printf("ERR:db:getPassword: %+v\n", err)
		return nil, err
	}
	pass.Password = string(password)
	jsonPass, err := json.Marshal(&pass)
	if err != nil {
		p.Logger.Printf("ERR:db:getPassword: %+v\n", err)
		return nil, err
	}
	resp.Payload = jsonPass
	return resp, nil
}

func (p *PG) getText(ctx context.Context, u *User, rec *Record) (*Record, error) {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Printf("ERR:db:getText: %+v\n", err)
		return nil, err
	}
	var text Text
	resp := &Record{Type: SRecordText}
	err = json.Unmarshal(rec.Payload, &text)
	if err != nil {
		p.Logger.Printf("ERR:db:getText: %+v\n", err)
		return nil, err
	}
	var encrypted pgtype.Bytea
	table := fmt.Sprintf("id%s.text", userID)
	sql := fmt.Sprintf(`select txt from %s where name=$1`, table)
	err = p.DB.QueryRow(ctx, sql, text.Name).Scan(&encrypted)
	if err != nil {
		p.Logger.Printf("ERR:db:getText: %+v\n", err)
		return nil, err
	}
	// Decrypt data
	txt, err := enc.Decrypt([]byte(p.Secret), encrypted.Bytes)
	if err != nil {
		p.Logger.Printf("ERR:db:getText: %+v\n", err)
		return nil, err
	}
	text.Text = string(txt)
	jsonText, err := json.Marshal(&text)
	if err != nil {
		p.Logger.Printf("ERR:svc:getText: %+v\n", err)
		return nil, err
	}
	resp.Payload = jsonText
	return resp, nil
}

func (p *PG) getCard(ctx context.Context, u *User, rec *Record) (*Record, error) {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Printf("ERR:db:getCard: %+v\n", err)
		return nil, err
	}
	var card Card
	resp := &Record{Type: SRecordCard}
	err = json.Unmarshal(rec.Payload, &card)
	if err != nil {
		p.Logger.Printf("ERR:db:getCard: %+v\n", err)
		return nil, err
	}
	var number, month, year string
	var encrypted pgtype.Bytea
	table := fmt.Sprintf("id%s.card", userID)
	sql := fmt.Sprintf(`select number, cvv, month, year from %s where name=$1`, table)
	err = p.DB.QueryRow(ctx, sql, card.Name).Scan(&number, &encrypted, &month, &year)
	if err != nil {
		p.Logger.Printf("ERR:db:getCard: %+v\n", err)
		return nil, err
	}
	// Decrypt data
	cvv, err := enc.Decrypt([]byte(p.Secret), encrypted.Bytes)
	if err != nil {
		p.Logger.Printf("ERR:db:getCard: %+v\n", err)
		return nil, err
	}
	card.Number = number
	card.Cvv = string(cvv)
	card.Month = month
	card.Year = year
	jsonCard, err := json.Marshal(&card)
	if err != nil {
		p.Logger.Printf("ERR:db:getCard: %+v\n", err)
		return nil, err
	}
	resp.Payload = jsonCard
	return resp, nil
}

func (p *PG) getBinary(ctx context.Context, u *User, rec *Record) (*Record, error) {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Printf("ERR:db:getBinary: %+v\n", err)
		return nil, err
	}
	var binary Binary
	resp := &Record{Type: SRecordBinary}
	err = json.Unmarshal(rec.Payload, &binary)
	if err != nil {
		p.Logger.Printf("ERR:db:getBinary: %+v\n", err)
		return nil, err
	}
	var encrypted pgtype.Bytea
	table := fmt.Sprintf("id%s.data", userID)
	sql := fmt.Sprintf(`select url from %s where name=$1`, table)
	err = p.DB.QueryRow(ctx, sql, binary.Name).Scan(&encrypted)
	if err != nil {
		p.Logger.Printf("ERR:db:getBinary: %+v\n", err)
		return nil, err
	}
	// Decrypt data
	path, err := enc.Decrypt([]byte(p.Secret), encrypted.Bytes)
	if err != nil {
		p.Logger.Printf("ERR:db:getBinary: %+v\n", err)
		return nil, err
	}
	binary.Path = string(path)
	jsonBinary, err := json.Marshal(&binary)
	if err != nil {
		p.Logger.Printf("ERR:db:getBinary: %+v\n", err)
		return nil, err
	}
	resp.Payload = jsonBinary
	return resp, nil
}

func (p *PG) Get(ctx context.Context, u *User, rec *Record) (*Record, error) {
	switch rec.Type {
	case SRecordPassword:
		return p.getPassword(ctx, u, rec)
	case SRecordText:
		return p.getText(ctx, u, rec)
	case SRecordCard:
		return p.getCard(ctx, u, rec)
	case SRecordBinary:
		return p.getBinary(ctx, u, rec)
	}
	unknown := &Record{Type: SRecordUnknown}
	return unknown, nil
}
