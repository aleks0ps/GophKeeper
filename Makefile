
SERVER := gophkeeper
CLIENT := gophclient

all: build

secret:
	tr -dc 'A-F0-9' < /dev/urandom | head -c32

build:
	go build -o ./cmd/gophkeeper/$(SERVER) ./cmd/gophkeeper
	go build -o ./cmd/client/$(CLIENT) ./cmd/client
clean:
	rm -v cmd/gophkeeper/$(SERVER)
	rm -v cmd/client/$(CLIENT)
