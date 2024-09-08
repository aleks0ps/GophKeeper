package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "APP: ", log.LstdFlags)
	logger.Printf("Here %s\n", "test")
}
