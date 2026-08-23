.PHONY: run test test-race lint fmt build docker-build tidy check

run:
	APP_ENV=development go run ./cmd/api

test:
	go test ./...

test-race:
	go test ./... -race

# Cobertura para mirar, no para poner un umbral: se prueban las reglas,
# no las líneas.
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	go vet ./...

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

docker-build:
	docker build -t salud-api:local .

# Lo que tiene que pasar antes de un commit
check: fmt test-race lint
