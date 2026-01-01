.PHONY: run swagger

debug:
	dlv debug ./cmd/api/main.go

swagger:
	swag init -g cmd/api/main.go -d . --parseInternal

go:
	go run ./cmd/api/main.go

run: swagger go
