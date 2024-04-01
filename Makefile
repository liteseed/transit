all:
	go mod tidy
	go build -o ./dist/transit ./build

gen-graphql:
	go get github.com/Khan/genqlient
	genqlient ./argraphql/genqlient.yaml

build-linux-bin:
	GOOS=linux GOARCH=amd64 go build -o ./build/transit ./cmd