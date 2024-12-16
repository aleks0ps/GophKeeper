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
	res := &Record{Type: SRecordUnknown}
	if rec.Type == SRecordPassword {
		var pass Password
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
			p.Logger.Println("svc:Get: ", err)
			return nil, err
		}
		res.Type = SRecordPassword
		res.Payload = jsonPass
	}
	return res, nil
}
