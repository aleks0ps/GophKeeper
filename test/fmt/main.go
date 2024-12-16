package main

import (
	"fmt"
	"strings"
)

func main() {
	id := "id1"
	tablePassword := `CREATE TABLE IF NOT EXISTS $1.password (
                           id BIGSERIAL PRIMARY KEY,
			   name VARCHAR(256) NOT NULL,
                           password VARCHAR(256) NOT NULL,
			   user_id INT,
			   CONSTRAINT fk_user
			     FOREIGN KEY (user_id)
			       REFERENCES users(id)
			       ON DELETE CASCADE
	                 );
			 CREATE UNIQUE INDEX $1_uniq_pass on $1.password (name)
			 `
	sqlTablePassword := strings.ReplaceAll(tablePassword, "$1", id)
	//sqlTablePassword := fmt.Sprintf(tablePassword, "id1")
	fmt.Println(sqlTablePassword)
}
