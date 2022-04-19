build:
	go build -o bin/main main.go

run:
	go run main.go

compile:
	GOOS=linux GOARCH=amd64 go build -o bin/oci-maintenance-reboot-amd64 main.go
	#GOOS=linux GOARCH=arm64 go build -o bin/oci-maintenance-reboot-arm64 main.go

all: compile