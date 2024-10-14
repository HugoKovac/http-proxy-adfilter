DB_TYPE = SQL_LITE

all: run

run:
	DB_TYPE=${DB_TYPE} go run cmd/filter/main.go -host 127.0.0.1

filter:
	DB_TYPE=${DB_TYPE} go run cmd/filter/main.go

build:
	go build -o ./bin/filter -ldflags="-w -s" -gcflags=all="-l -B -wb=false" cmd/filter/main.go 

build_migrate:
	go build -o ./bin/migrate cmd/migration/main.go

migrate:
	DB_TYPE=${DB_TYPE} go run cmd/migration/main.go

display_db:
	DB_TYPE=${DB_TYPE} go run cmd/display_db/main.go


glinet:
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -gcflags=all="-l -B -wb=false" -o ./bin/filter_glinet cmd/filter/main.go 
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -gcflags=all="-l -B -wb=false" -o ./bin/migrate_glinet cmd/migration/main.go

docker: glinet
	docker compose up --no-deps --force-recreate --build