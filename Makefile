BIN_DIR=bin
APP_NAME=migrate

all: clean linux macos windows

linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-linux main.go

macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) main.go

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME).exe main.go

clean:
	rm -rf $(BIN_DIR)/* 