package db

import (
	"context"
	"encoding/json"
	"fmt"
)

func (p *PG) List(ctx context.Context, u *User) ([]Record, error) {
	var recs []Record
	var userID string
	// find user id
	err := p.DB.QueryRow(ctx, `SELECT id from users WHERE login=$1`, u.ID).Scan(&userID)
	if err != nil {
		p.Logger.Println(err)
		return nil, err
	}
	var ID, name string
	// find all passwords
	sql := fmt.Sprintf("SELECT id, name FROM id%s.password WHERE user_id=%[1]s", userID)
	rows, err := p.DB.Query(ctx, sql)
	if err != nil {
		p.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&ID, &name)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		payload, err := json.Marshal(Password{Name: name})
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		recs = append(recs, Record{Type: SRecordPassword, Payload: payload})
	}
	if err := rows.Err(); err != nil {
		p.Logger.Println(err)
		return recs, err
	}
	// find all binary
	sql = fmt.Sprintf("SELECT id, name FROM id%s.data WHERE user_id=%[1]s", userID)
	rows, err = p.DB.Query(ctx, sql)
	if err != nil {
		p.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&ID, &name)
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		payload, err := json.Marshal(Binary{Name: name})
		if err != nil {
			p.Logger.Println(err)
			return nil, err
		}
		recs = append(recs, Record{Type: SRecordBinary, Payload: payload})
	}
	if err := rows.Err(); err != nil {
		p.Logger.Println(err)
		return recs, err
	}
	return recs, nil
}
