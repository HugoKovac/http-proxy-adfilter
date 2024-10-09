DB_TYPE = SQL_LITE

all: run

run:
	DB_TYPE=${DB_TYPE} go run cmd/filter/main.go -pem cert.pem -key key.pem

filter:
	DB_TYPE=${DB_TYPE} go run cmd/filter/main.go

build:
	go build -o ./bin/filter cmd/filter/main.go

build_migrate:
	go build -o ./bin/migrate cmd/migration/main.go

migrate:
	DB_TYPE=${DB_TYPE} go run cmd/migration/main.go

display_db:
	DB_TYPE=${DB_TYPE} go run cmd/display_db/main.go

db_up:
	docker compose up -d

db_down:
	docker compose down	

glinet:
	GOOS=linux GOARCH=arm GOARM=7 go build -o ./bin/filter_glinet cmd/filter/main.go 
	GOOS=linux GOARCH=arm GOARM=7 go build -o ./bin/migrate_glinet cmd/migration/main.go
