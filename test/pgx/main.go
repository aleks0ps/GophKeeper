package main

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func main() {
	id := pgx.Identifier{"1"}
	fmt.Println("VAR: ", id.Sanitize())
}
