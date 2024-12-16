package db

import (
	"context"
	"encoding/json"
	"fmt"
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
		_, err = p.DB.Exec(ctx, sql, pass.Name, pass.Password, userID)
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
		sql := fmt.Sprintf(`insert into %s(name,url,user_id) values ($1,$2,$3)`, table)
		_, err = p.DB.Exec(ctx, sql, binary.Name, binary.Path, userID)
		if err != nil {
			p.Logger.Println(err)
			return err
		}
	}
	return nil
}
