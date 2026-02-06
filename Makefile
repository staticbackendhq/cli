build:
	@go build -o backend

test:
	@go test -cover ./...

pkg: build
	@rm -rf dist/*
	@echo "building linux binaries"
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o dist/linux-amd64-backend
	@CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o dist/linux-arm64-backend
	@CGO_ENABLED=0 GOARCH=386 GOOS=linux go build -o dist/linux-386-backend
	@CGO_ENABLED=0 GOARCH=arm GOOS=linux GOARM=7 go build -o dist/linux-arm-backend
	@echo "building mac binaries"
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o dist/darwin-amd64-backend
	@CGO_ENABLED=0 GOARCH=arm64 GOOS=darwin go build -o dist/darwin-arm64-backend
	@echo "building windows binaries"
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o dist/windows-amd64-backend.exe
	@CGO_ENABLED=0 GOARCH=386 GOOS=windows go build -o dist/windows-386-backend.exe

compress:
	@echo "compressing binaries"
	@gzip dist/*