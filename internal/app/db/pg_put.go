package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aleks0ps/GophKeeper/internal/app/enc"
	creditcard "github.com/durango/go-credit-card"
)

func (p *PG) Put(ctx context.Context, u *User, rec *Record) error {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Println(err)
		return err
	}
	if rec.Type == SRecordPassword {
		var pass Password
		err := json.Unmarshal(rec.Payload, &pass)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
		table := fmt.Sprintf("id%s.password", userID)
		sql := fmt.Sprintf(`insert into %s(name,password,user_id) values ($1, $2, $3)`, table)
		// encrypt data
		encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(pass.Password))
		if err != nil {
			p.Logger.Printf("ERR:db:put %s\n", err)
			return err
		}
		_, err = p.DB.Exec(ctx, sql, pass.Name, string(encrypted), userID)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
	} else if rec.Type == SRecordBinary {
		var binary Binary
		err := json.Unmarshal(rec.Payload, &binary)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
		table := fmt.Sprintf("id%s.data", userID)
		// encrypt data
		encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(binary.Path))
		if err != nil {
			p.Logger.Println(err)
			return err
		}
		sql := fmt.Sprintf(`insert into %s(name,url,user_id) values ($1,$2,$3)`, table)
		_, err = p.DB.Exec(ctx, sql, binary.Name, encrypted, userID)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
	} else if rec.Type == SRecordCard {
		var card Card
		err := json.Unmarshal(rec.Payload, &card)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
		// validate card
		cc := creditcard.Card{Number: card.Number, Cvv: card.Cvv, Month: card.Month, Year: card.Year}
		// allow test cards
		err = cc.Validate(true)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
		table := fmt.Sprintf("id%s.card", userID)
		sql := fmt.Sprintf(`insert into %s(name,number,cvv,month,year,user_id) values ($1,$2,$3,$4,$5,$6)`, table)
		// encrypt data
		encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(card.Cvv))
		if err != nil {
			p.Logger.Printf("ERR:db:put: %+v\n", err)
			return err
		}
		_, err = p.DB.Exec(ctx, sql, card.Name, card.Number, encrypted, card.Month, card.Year, userID)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
	} else if rec.Type == SRecordText {
		var text Text
		err := json.Unmarshal(rec.Payload, &text)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
		table := fmt.Sprintf("id%s.text", userID)
		sql := fmt.Sprintf(`insert into %s(name,txt,user_id) values ($1,$2,$3)`, table)
		// encrypt data
		encrypted, err := enc.Encrypt([]byte(p.Secret), []byte(text.Text))
		if err != nil {
			p.Logger.Println(err)
			return err
		}
		_, err = p.DB.Exec(ctx, sql, text.Name, encrypted, userID)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
	}
	return nil
}
