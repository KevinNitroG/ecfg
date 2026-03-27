.PHONY: build clean test vet install

build: ecfg-lsp

ecfg-lsp:
	go build -o ecfg-lsp ./cmd/ecfg-lsp

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o ecfg-lsp-darwin-amd64 ./cmd/ecfg-lsp
	GOOS=darwin GOARCH=arm64 go build -o ecfg-lsp-darwin-arm64 ./cmd/ecfg-lsp

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ecfg-lsp-linux-amd64 ./cmd/ecfg-lsp
	GOOS=linux GOARCH=arm64 go build -o ecfg-lsp-linux-arm64 ./cmd/ecfg-lsp

build-windows:
	GOOS=windows GOARCH=amd64 go build -o ecfg-lsp-windows-amd64.exe ./cmd/ecfg-lsp

build-all: build-darwin build-linux build-windows

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -f ecfg-lsp ecfg-lsp-*

install:
	go install ./cmd/ecfg-lsp