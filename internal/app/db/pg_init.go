package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	myerror "github.com/aleks0ps/GophKeeper/internal/app/error"
	"github.com/aleks0ps/GophKeeper/internal/app/util"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func initDB(ctx context.Context, db *pgxpool.Pool, logger *log.Logger) {
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
				  id BIGSERIAL PRIMARY KEY,
				  login TEXT NOT NULL,
				  password TEXT NOT NULL
				);
				CREATE UNIQUE INDEX users_uniq_login ON users (login);
				`)
	if err != nil {
		logger.Printf("cannot init db: %+v\n", err)
	}
}

func NewDB(ctx context.Context, DSN string, logger *log.Logger, secret string) (*PG, error) {
	poolConfig, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		logger.Printf("ERR: %+v\n", err)
		return nil, err
	}
	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Printf("ERR: %+v\n", err)
		return nil, err
	}
	initDB(ctx, db, logger)
	return &PG{DB: db, Logger: logger, Secret: secret}, nil
}

func (p *PG) initSchema(ctx context.Context, u *User) error {
	// per-user schema
	sqlSchemaName := pgx.Identifier{kSchemaPrefix + u.ID}.Sanitize()
	// trim doubel qoutes
	sqlSchemaName = strings.Trim(sqlSchemaName, "\"")
	sqlSchema := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", sqlSchemaName)
	p.Logger.Println("INFO:initSchema: ", sqlSchema)
	p.Logger.Println("INFO initSchema: ", sqlSchemaName)
	if _, err := p.DB.Exec(ctx, sqlSchema); err != nil {
		p.Logger.Println("ERR: initSchema: ", err)
		return err
	}
	// create password table in schema
	tablePassword := `CREATE TABLE IF NOT EXISTS %s.password (
                           id BIGSERIAL PRIMARY KEY,
			   name TEXT NOT NULL,
                           password BYTEA NOT NULL,
			   user_id INT NOT NULL,
			   CONSTRAINT fk_user
			     FOREIGN KEY (user_id)
			       REFERENCES users(id)
			       ON DELETE CASCADE
	                 );
			 CREATE UNIQUE INDEX %[1]s_uniq_pass on %[1]s.password (name)
			 `
	sqlTablePassword := fmt.Sprintf(tablePassword, sqlSchemaName)
	p.Logger.Println("INFO:initSchema: ", sqlTablePassword)
	if _, err := p.DB.Exec(ctx, sqlTablePassword); err != nil {
		p.Logger.Println("ERR:initSchema: ", err)
		return err
	}
	// create card table in schema
	tableCard := `CREATE TABLE IF NOT EXISTS %s.card (
                         id BIGSERIAL PRIMARY KEY,
			 name TEXT NOT NULL,
			 number TEXT NOT NULL,
			 cvv BYTEA NOT NULL,
			 month INT NOT NULL,
			 year INT NOT NULL,
			 user_id INT NOT NULL,
			 CONSTRAINT fk_user
			   FOREIGN KEY (user_id)
			     REFERENCES users(id)
			     ON DELETE CASCADE
                      );
		      CREATE UNIQUE INDEX %[1]s_uniq_card on %[1]s.card (name)
		      `
	sqlTableCard := fmt.Sprintf(tableCard, sqlSchemaName)
	p.Logger.Println("INFO:initSchema: ", sqlTableCard)
	if _, err := p.DB.Exec(ctx, sqlTableCard); err != nil {
		p.Logger.Println("ERR:initSchema: ", err)
		return err
	}
	// create data/binary table
	tableData := `CREATE TABLE IF NOT EXISTS %s.data (
		        id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			url BYTEA NOT NULL,
			user_id INT NOT NULL,
		        CONSTRAINT fk_user
			  FOREIGN KEY (user_id)
			    REFERENCES users(id)
			    ON DELETE CASCADE
	              );
		      CREATE UNIQUE INDEX %[1]s_uniq_data on %[1]s.data (name)
		      `
	sqlTableData := fmt.Sprintf(tableData, sqlSchemaName)
	p.Logger.Println("INFO:initSchema: ", sqlTableData)
	if _, err := p.DB.Exec(ctx, sqlTableData); err != nil {
		p.Logger.Println("ERR:initSchema: ", err)
		return err
	}
	// create text table
	tableText := `CREATE TABLE IF NOT EXISTS %s.text (
		        id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			txt BYTEA NOT NULL,
			user_id INT NOT NULL,
		        CONSTRAINT fk_user
			  FOREIGN KEY (user_id)
			    REFERENCES users(id)
			    ON DELETE CASCADE
	              );
		      CREATE UNIQUE INDEX %[1]s_uniq_text on %[1]s.text (name)
		      `
	sqlTableText := fmt.Sprintf(tableText, sqlSchemaName)
	p.Logger.Println("INFO:initSchema: ", sqlTableText)
	if _, err := p.DB.Exec(ctx, sqlTableText); err != nil {
		p.Logger.Println("ERR:initSchema: ", err)
		return err
	}
	return nil
}

func (p *PG) Register(ctx context.Context, u *User) error {
	hPassword, err := util.Hash(u.Password)
	if err != nil {
		p.Logger.Println("ERR:Register: ", err)
		return err
	}
	if _, err := p.DB.Exec(ctx, `INSERT INTO users(login, password) values ($1, $2)`, u.Login, hPassword); err != nil {
		p.Logger.Println("ERR:Register: ", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Record already exists
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return myerror.ErrLoginAlreadyTaken
			}
		}
		return err
	}
	var ID string
	err = p.DB.QueryRow(ctx, `SELECT id FROM users WHERE login=$1`, u.Login).Scan(&ID)
	if err != nil {
		p.Logger.Println("ERR:Register: ", err)
		return err
	}
	u.ID = ID
	// Create schema per user
	if err = p.initSchema(ctx, u); err != nil {
		p.Logger.Println("ERR:Register: ", err)
		return err
	}
	return nil
}

func (p *PG) Login(ctx context.Context, u *User) error {
	var hPassword string
	err := p.DB.QueryRow(ctx, `SELECT password FROM users WHERE login=$1`, u.Login).Scan(&hPassword)
	if err != nil {
		p.Logger.Println("ERR:Login: ", err)
		return err
	}
	err = util.CheckPasswordHash(hPassword, u.Password)
	if err != nil {
		p.Logger.Println("ERR:Login: ", err)
		return err
	}
	return nil
}
