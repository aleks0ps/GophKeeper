package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aleks0ps/GophKeeper/internal/app/enc"
	creditcard "github.com/durango/go-credit-card"
)

func (p *PG) putPassword(ctx context.Context, u *User, rec *Record) error {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Println(err)
		return err
	}
	var pass Password
	err = json.Unmarshal(rec.Payload, &pass)
	if err != nil {
		p.Logger.Printf("ERR:db:putPassword:jsonUnmarshal: %+v\n", err)
		return err
	}
	table := fmt.Sprintf("id%s.password", userID)
	sql := fmt.Sprintf(`insert into %s(name,password,user_id) values ($1, $2, $3)`, table)
	// encrypt data
	encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(pass.Password))
	if err != nil {
		p.Logger.Printf("ERR:db:putPassword:enc.Encrypt: %s\n", err)
		return err
	}
	_, err = p.DB.Exec(ctx, sql, pass.Name, encrypted, userID)
	if err != nil {
		p.Logger.Printf("ERR:db:putPassword:p.DB.Exec: %+v\n", err)
		return err
	}
	return nil
}

func (p *PG) putBinary(ctx context.Context, u *User, rec *Record) error {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Printf("ERR:db:putBinary: %+v\n", err)
		return err
	}
	var binary Binary
	err = json.Unmarshal(rec.Payload, &binary)
	if err != nil {
		p.Logger.Printf("ERR:db:putBinary: %+v\n", err)
		return err
	}
	table := fmt.Sprintf("id%s.data", userID)
	// encrypt data
	encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(binary.Path))
	if err != nil {
		p.Logger.Printf("ERR:db:putBinary: %+v\n", err)
		return err
	}
	sql := fmt.Sprintf(`insert into %s(name,url,user_id) values ($1,$2,$3)`, table)
	_, err = p.DB.Exec(ctx, sql, binary.Name, encrypted, userID)
	if err != nil {
		p.Logger.Printf("ERR:db:putBinary: %+v\n", err)
		return err
	}
	return nil
}

func (p *PG) putCard(ctx context.Context, u *User, rec *Record) error {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Printf("ERR:db:putCard: %+v\n", err)
		return err
	}
	var card Card
	err = json.Unmarshal(rec.Payload, &card)
	if err != nil {
		p.Logger.Printf("ERR:db:putCard: %+v\n", err)
		return err
	}
	// validate card
	cc := creditcard.Card{Number: card.Number, Cvv: card.Cvv, Month: card.Month, Year: card.Year}
	// allow test cards
	err = cc.Validate(true)
	if err != nil {
		p.Logger.Printf("ERR:db:putCard: %+v\n", err)
		return err
	}
	table := fmt.Sprintf("id%s.card", userID)
	sql := fmt.Sprintf(`insert into %s(name,number,cvv,month,year,user_id) values ($1,$2,$3,$4,$5,$6)`, table)
	// encrypt data
	encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(card.Cvv))
	if err != nil {
		p.Logger.Printf("ERR:db:putCard: %+v\n", err)
		return err
	}
	_, err = p.DB.Exec(ctx, sql, card.Name, card.Number, encrypted, card.Month, card.Year, userID)
	if err != nil {
		p.Logger.Printf("ERR:db:putCard: %+v\n", err)
		return err
	}
	return nil
}

func (p *PG) putText(ctx context.Context, u *User, rec *Record) error {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Printf("ERR:db:putText: %+v\n", err)
		return err
	}
	var text Text
	err = json.Unmarshal(rec.Payload, &text)
	if err != nil {
		p.Logger.Printf("ERR:db:putText: %+v\n", err)
		return err
	}
	table := fmt.Sprintf("id%s.text", userID)
	sql := fmt.Sprintf(`insert into %s(name,txt,user_id) values ($1,$2,$3)`, table)
	// encrypt data
	encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(text.Text))
	if err != nil {
		p.Logger.Printf("ERR:db:putText: %+v\n", err)
		return err
	}
	_, err = p.DB.Exec(ctx, sql, text.Name, encrypted, userID)
	if err != nil {
		p.Logger.Printf("ERR:db:putText: %+v\n", err)
		return err
	}
	return nil
}

func (p *PG) Put(ctx context.Context, u *User, rec *Record) error {
	switch rec.Type {
	case SRecordPassword:
		return p.putPassword(ctx, u, rec)
	case SRecordBinary:
		return p.putBinary(ctx, u, rec)
	case SRecordCard:
		return p.putCard(ctx, u, rec)
	case SRecordText:
		return p.putText(ctx, u, rec)
	}
	return nil
}
