.PHONY: run create-mig migrate-up migrate-down compose-up compose-down

# Run locally (outside of Docker)
run:
	go run main.go

create-mig:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

migrate-up:
	migrate -path internal/db/migrations -database $(DB_URL) up

migrate-down:
	migrate -path internal/db/migrations -database $(DB_URL) down

# Docker Compose
build:
	docker-compose up --build
compose-up:
	docker-compose up
compose-down:
	docker-compose down
