build:
	go build -o bin/tbm src/main.go

run:
	go run src/main.go help

test_switch:
	go run src/main.go switch test

test_rename:
	go run src/main.go rename test

help:
	go run src/main.go help

compile:
	# FreeBDS
	GOOS=freebsd GOARCH=amd64 go build -o bin/tbm-freebsd-amd64 src/main.go
	# MacOS
	GOOS=darwin GOARCH=amd64 go build -o bin/tbm-darwin-amd64 src/main.go
	# Linux
	GOOS=linux GOARCH=amd64 go build -o bin/tbm-linux-amd64 src/main.go
	# Windows
	GOOS=windows GOARCH=amd64 go build -o bin/tbm-windows-amd64 src/main.go
