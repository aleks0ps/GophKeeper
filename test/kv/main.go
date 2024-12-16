package main

import (
	"fmt"
	"strings"
)

func main() {
	opt := strings.Split("-login=user", "=")
	fmt.Printf("%v \n", opt[1])
}
