all: build

build:
	go build -o bin/rat rat.go
	# Windows
	GOOS=windows GOARCH=amd64 go build -o ../deploy/files/rat.exe rat.go
run:
	go run rat.go
