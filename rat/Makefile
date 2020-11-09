all: build

build:
	go build -o bin/rat
	# Windows
	GOOS=windows GOARCH=amd64 go build -o ../deploy/files/rat.exe
run:
	go run .
