GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test

SERVER_BIN=redis_server
CLI_BIN=redis_cli

all: server cli
server:
	$(GO_BUILD) -o $(SERVER_BIN) $(shell pwd)/internal/app/server
cli:
	$(GO_BUILD) -o $(CLI_BIN) $(shell pwd)/internal/app/cli
test:
	$(GO_TEST) ./...
clean:
	find . -name *.aof -delete
	find . -name *.rcl* -delete
	find . -name *.rdb -delete
	find . -name redis_server -delete
	find . -name redis_cli -delete
