.PHONY: run swagger

swagger:
	swag init -g cmd/api/main.go -d . --parseInternal

go:
	go run ./cmd/api/main.go

run: swagger go
