GOPHKEEPER := ./cmd/gophkeeper
BIN := gophkeeper

all: build

build:
	go build -o $(GOPHKEEPER)/$(BIN) $(GOPHKEEPER)
