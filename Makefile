.PHONY: run swagger test lint build docker-up docker-down migrate-up migrate-down migrate-version migrate-force

debug:
	dlv debug ./cmd/api/main.go

swagger:
	swag init -g cmd/api/main.go -d . --parseInternal --parseDependency

go:
	go run ./cmd/api/main.go

run: swagger go

test:
	go test ./... -v -count=1

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/api/main.go

build-migrate:
	CGO_ENABLED=0 go build -o bin/migrate ./cmd/migrate/main.go

migrate-up: build-migrate
	./bin/migrate up

migrate-down: build-migrate
	./bin/migrate down

migrate-version: build-migrate
	./bin/migrate version

migrate-force: build-migrate
	./bin/migrate force $(VERSION)

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
