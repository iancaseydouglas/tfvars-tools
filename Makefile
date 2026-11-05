.PHONY: all build test vet fmt check cross clean

all: check build

build:
	go build -o bin/tfv ./cmd/tfv

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

check: vet test
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

cross:
	GOOS=linux GOARCH=arm64 go build -o /dev/null ./cmd/tfv
	GOOS=darwin GOARCH=arm64 go build -o /dev/null ./cmd/tfv
	GOOS=windows GOARCH=amd64 go build -o /dev/null ./cmd/tfv

clean:
	rm -rf bin dist
