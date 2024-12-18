package db

import (
	"context"
	"encoding/json"
	"fmt"
)

func (p *PG) Get(ctx context.Context, u *User, rec *Record) (*Record, error) {
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Println(err)
		return nil, err
	}
	if rec.Type == SRecordPassword {
		var pass Password
		resp := &Record{Type: SRecordPassword}
		err := json.Unmarshal(rec.Payload, &pass)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		var password string
		table := fmt.Sprintf("id%s.password", userID)
		sql := fmt.Sprintf(`select password from %s where name=$1`, table)
		err = p.DB.QueryRow(ctx, sql, pass.Name).Scan(&password)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		pass.Password = password
		jsonPass, err := json.Marshal(&pass)
		if err != nil {
			p.Logger.Println("ERR:db:Get: ", err)
			return nil, err
		}
		resp.Payload = jsonPass
		return resp, nil
	} else if rec.Type == SRecordText {
		var text Text
		resp := &Record{Type: SRecordText}
		err := json.Unmarshal(rec.Payload, &text)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		var txt string
		table := fmt.Sprintf("id%s.text", userID)
		sql := fmt.Sprintf(`select txt from %s where name=$1`, table)
		err = p.DB.QueryRow(ctx, sql, text.Name).Scan(&txt)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		text.Text = txt
		jsonText, err := json.Marshal(&text)
		if err != nil {
			p.Logger.Println("ERR:svc:Get: ", err)
			return nil, err
		}
		resp.Payload = jsonText
		return resp, nil
	} else if rec.Type == SRecordCard {
		var card Card
		resp := &Record{Type: SRecordCard}
		err := json.Unmarshal(rec.Payload, &card)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		var number, cvv, month, year string
		table := fmt.Sprintf("id%s.card", userID)
		sql := fmt.Sprintf(`select number, cvv, month, year from %s where name=$1`, table)
		err = p.DB.QueryRow(ctx, sql, card.Name).Scan(&number, &cvv, &month, &year)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		card.Number = number
		card.Cvv = cvv
		card.Month = month
		card.Year = year
		jsonCard, err := json.Marshal(&card)
		if err != nil {
			p.Logger.Println("svc:Get: ", err)
			return nil, err
		}
		resp.Payload = jsonCard
		return resp, nil
	} else if rec.Type == SRecordBinary {
		var binary Binary
		resp := &Record{Type: SRecordBinary}
		err := json.Unmarshal(rec.Payload, &binary)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		var path string
		table := fmt.Sprintf("id%s.data", userID)
		sql := fmt.Sprintf(`select url from %s where name=$1`, table)
		err = p.DB.QueryRow(ctx, sql, binary.Name).Scan(&path)
		if err != nil {
			p.Logger.Println("ERR:db:GET: ", err)
			return nil, err
		}
		binary.Path = path
		jsonBinary, err := json.Marshal(&binary)
		if err != nil {
			p.Logger.Println("svc:Get: ", err)
			return nil, err
		}
		resp.Payload = jsonBinary
		return resp, nil
	}
	unknown := &Record{Type: SRecordUnknown}
	return unknown, nil
}
