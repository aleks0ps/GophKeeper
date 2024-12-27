SERVER := gophkeeper
CLIENT := gophclient

DATE=$(shell date)
GOVERSION=$(shell go version)
VERSION=$(shell git describe --tags --abbrev=8 --dirty --always --long)

PREFIX="github.com/aleks0ps/GophKeeper/cmd/client/version"
LDFLAGS=
LDFLAGS+= -X '$(PREFIX).Version=$(VERSION)'
LDFLAGS+= -X '$(PREFIX).Date=$(DATE)'
LDFLAGS+= -X '$(PREFIX).GoVersion=$(GOVERSION)'

all: build

secret:
	tr -dc 'A-F0-9' < /dev/urandom | head -c32

lint:
	golangci-lint run ./...
	staticcheck -checks all ./...

.PHONY: test
test:
	go test -v ./internal/...

build:
	go build -o ./cmd/gophkeeper/$(SERVER) ./cmd/gophkeeper
	go build -ldflags "$(LDFLAGS)" -o ./cmd/client/$(CLIENT) ./cmd/client
	env GOOS=windows GOARCH=arm64 go build -o ./cmd/gophkeeper/$(SERVER)-win ./cmd/gophkeeper
	env GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ./cmd/client/$(CLIENT)-win ./cmd/client
	env GOOS=darwin GOARCH=arm64 go build -o ./cmd/gophkeeper/$(SERVER)-mac ./cmd/gophkeeper
	env GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ./cmd/client/$(CLIENT)-mac ./cmd/client
	
clean:
	rm -vf cmd/gophkeeper/$(SERVER)-win
	rm -rf cmd/gophkeeper/$(SERVER)-mac
	rm -rf cmd/gophkeeper/$(SERVER)
	rm -vf cmd/client/$(CLIENT)-win
	rm -vf cmd/client/$(CLIENT)-mac
	rm -vf cmd/client/$(CLIENT)
	rm -vf key.pem cert.pem
