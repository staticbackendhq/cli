build:
	@go build -o backend

test:
	@go test -cover ./...

pkg: build
	@rm -rf dist/*
	@echo "building linux binaries"
	@GOARCH=amd64 GOOS=linux go build -o dist/binary-for-linux-64-bit
	@GOARCH=386 GOOS=linux go build -o dist/binary-for-linux-32-bit
	@echo "building mac binaries"
	@GOARCH=amd64 GOOS=darwin go build -o dist/binary-for-mac-64-bit
	@GOARCH=386 GOOS=darwin go build -o dist/binary-for-mac-32-bit
	@echo "building windows binaries"
	@GOARCH=amd64 GOOS=windows go build -o dist/binary-for-windows-64-bit.exe
	@GOARCH=386 GOOS=windows go build -o dist/binary-for-windows-32-bit.exe
	@echo "compressing binaries"
	@gzip dist/*