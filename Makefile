.PHONY: build run clean

APP_NAME ?= family-score
BIN_DIR ?= bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)
ADDR ?= 127.0.0.1:8080
DATA_DIR ?= $(HOME)/.family-score

build:
	@mkdir -p $(BIN_DIR)
	go mod download
	go build -trimpath -ldflags="-s -w" -o $(BIN_PATH) .

run: build
	FAMILY_SCORE_ADDR=$(ADDR) FAMILY_SCORE_DATA_DIR=$(DATA_DIR) ./$(BIN_PATH)

clean:
	rm -rf $(BIN_DIR)
