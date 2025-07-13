.PHONY: run create-mig migrate-up migrate-down 

run:
	go run main.go

create-mig:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

migrate-up:
	migrate -path internal/db/migrations -database $(DB_URL) up

migrate-down:
	migrate -path internal/db/migrations -database $(DB_URL) down

buildbin:
	mkdir -p .bin
	go build -o .bin/myapp
