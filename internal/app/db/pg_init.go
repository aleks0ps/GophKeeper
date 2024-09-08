package db

import (
	"context"
	"errors"
	"fmt"
	"log"

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
		logger.Println("cannot init db: ", err)
	}
}

func NewDB(ctx context.Context, DSN string, logger *log.Logger) (*PG, error) {
	poolConfig, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		logger.Println("ERR: ", err)
		return nil, err
	}
	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Println("ERR: ", err)
		return nil, err
	}
	initDB(ctx, db, logger)
	return &PG{DB: db, Logger: logger}, nil
}

func (p *PG) initSchema(ctx context.Context, u *User) error {
	// per-user schema
	sqlSchemaName := pgx.Identifier{kSchemaPrefix + u.ID}.Sanitize()
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
                           password VARCHAR(256),
			   user_id INT,
			   CONSTRAINT fk_user
			     FOREIGN KEY (user_id)
			       REFERENCES users(id)
			       ON DELETE CASCADE
	                 )`
	sqlTablePassword := fmt.Sprintf(tablePassword, sqlSchemaName)
	//p.Logger.Println("INFO:initSchema: ", sqlTablePassword)
	if _, err := p.DB.Exec(ctx, sqlTablePassword); err != nil {
		p.Logger.Println("ERR:initSchema: ", err)
		return err
	}
	// create card table in schema
	tableCard := `CREATE TABLE IF NOT EXISTS %s.card (
                         id BIGSERIAL PRIMARY KEY,
			 key VARCHAR(100),
			 cvc INT,
			 valid DATE NOT NULL,
			 number INT,
			 user_id INT,
			 CONSTRAINT fk_user
			   FOREIGN KEY (user_id)
			     REFERENCES users(id)
			     ON DELETE CASCADE
                      )`
	sqlTableCard := fmt.Sprintf(tableCard, sqlSchemaName)
	//p.Logger.Println("INFO:initSchema: ", sqlTableCard)
	if _, err := p.DB.Exec(ctx, sqlTableCard); err != nil {
		p.Logger.Println("ERR:initSchema: ", err)
		return err
	}
	// create
	tableData := `CREATE TABLE IF NOT EXISTS %s.data (
		        id BIGSERIAL PRIMARY KEY,
			is_text BOOLEAN DEFAULT 'f',
			is_binary BOOLEAN DEFAULT 't',
			url TEXT NOT NULL,
			user_id INT,
		        CONSTRAINT fk_user
			  FOREIGN KEY (user_id)
			    REFERENCES users(id)
			    ON DELETE CASCADE
	              )`
	sqlTableData := fmt.Sprintf(tableData, sqlSchemaName)
	//p.Logger.Println("INFO:initSchema: ", sqlTableData)
	if _, err := p.DB.Exec(ctx, sqlTableData); err != nil {
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
