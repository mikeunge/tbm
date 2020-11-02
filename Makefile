build:
	go build -o bin/tbm main.go

compile:
	# FreeBDS
	GOOS=freebsd GOARCH=amd64 go build -o bin/tbm-freebsd-amd64 main.go
	# MacOS
	GOOS=darwin GOARCH=amd64 go build -o bin/tbm-darwin-amd64 main.go
	# Linux
	GOOS=linux GOARCH=amd64 go build -o bin/tbm-linux-amd64 main.go
	# Windows
	GOOS=windows GOARCH=amd64 go build -o bin/tbm-windows-amd64 main.go
