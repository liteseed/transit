BINARY_NAME := api
PKG := github.com/liteseed/api
VERSION := 0.0.4

dev:
	go mod tidy
	go build -o ./build/dev/${BINARY_NAME} -ldflags="-X main.Version=${VERSION}-dev"  ./cmd/main.go

release:
	go mod tidy
	GOARCH=386 GOOS=linux go build -o ./build/release/${BINARY_NAME}-linux-386 -ldflags="-X main.Version=${VERSION}" ./cmd/main.go

clean:
	go clean
	rm -rf ./build
