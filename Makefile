SERVER := gophkeeper
CLIENT := gophclient

all: build

secret:
	tr -dc 'A-F0-9' < /dev/urandom | head -c32

build:
	go build -o ./cmd/gophkeeper/$(SERVER) ./cmd/gophkeeper
	go build -o ./cmd/client/$(CLIENT) ./cmd/client
	env GOOS=windows GOARCH=arm64 go build -o ./cmd/gophkeeper/$(SERVER)-win ./cmd/gophkeeper
	env GOOS=windows GOARCH=arm64 go build -o ./cmd/client/$(CLIENT)-win ./cmd/client
	env GOOS=ios GOARCH=arm64  go build -o ./cmd/gophkeeper/$(SERVER)-mac ./cmd/gophkeeper
	env GOOS=ios GOARCH=arm64  go build -o ./cmd/client/$(CLIENT)-mac ./cmd/client
	
clean:
	rm -v cmd/gophkeeper/$(SERVER)
	rm -v cmd/client/$(CLIENT)
