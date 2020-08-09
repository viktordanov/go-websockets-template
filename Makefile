BINARY_NAME=go-youtube-sync

all: compile run

compile:
	go build -o $(BINARY_NAME) ./cmd/cli

run:
	./$(BINARY_NAME)