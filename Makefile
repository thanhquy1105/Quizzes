.PHONY: run build test clean setup help

## run: Run the quiz server
run:
	go run cmd/quizserver/main.go

## build: Build the quiz server binary
build:
	go build -o quizserver cmd/quizserver/main.go

## test: Run all tests
test:
	go test -v ./...

## docker-up: Start full stack with Docker (builds images)
docker-up:
	docker compose up -d --build

## docker-down: Stop Redis and MySQL containers
docker-down:
	docker compose down

## docker-logs: View Docker logs
docker-logs:
	docker compose logs -f

## clean: Remove binary and log files
clean:
	rm -f quizserver
	rm -f *.log

## setup: Install dependencies
setup:
	go mod tidy

## help: Show this help message
help:
	@echo "Usage: make [target] [options]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' -e 's/: /	/g' | column -t -s '	' | sed 's/^/  /'
