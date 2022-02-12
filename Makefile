buildBot:
	go build -o tgBot cmd/bot/main.go

buildServer:
	go build -o server cmd/server/main.go

build: download lint  buildBot buildServer

runBot: buildBot
	./tgBot

runServer: buildServer
	./server

download:
	go mod download

lint:
	golangci-lint run

.PHONY: download lint runBot runServer buildBot buildServer build