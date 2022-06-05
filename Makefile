build:
	@go build -o backend

test:
	@go test -cover ./...

pkg: build
	@rm -rf dist/*
	@echo "building linux binaries"
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o dist/linux-amd64-backend
	@echo "building mac binaries"
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o dist/darwin-amd64-backend
	@echo "building windows binaries"
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o dist/windows-amd64-backend.exe

compress:
	@echo "compressing binaries"
	@gzip dist/*