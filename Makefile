include .env
export $(shell sed 's/=.*//' .env)

PROJ_NAME = manipulator

MAIN_PATH = cmd/main.go
BUILD_PATH = build/package/

.DEFAULT_GOAL := run

run:
	go run $(MAIN_PATH)

build: clean
	go build --ldflags '-extldflags "-static"' -v -o $(BUILD_PATH)$(PROJ_NAME) $(MAIN_PATH)

clean:
	rm -rf $(BUILD_PATH)*

tests:
	go test ./...

lint:
	golangci-lint run

cloc:
	cloc . --exclude-ext=yml,mod,sum,xml