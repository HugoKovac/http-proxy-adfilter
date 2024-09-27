all: run

run:
	go run cmd/filter/main.go

filter:
	go run cmd/filter/main.go

build:
	go build -o ./bin/filter cmd/filter/main.go

migrate:
	go run cmd/migration/main.go

db_up:
	docker compose up -d

db_down:
	docker compose down	