include .env

run:
	go run ./cmd/main.go

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

migrations-up:
	cd .\db\migrations\ && goose postgres "host=${PG_HOST} port=${PG_PORT} user=${PG_LOGIN} database=${PG_NAME} password=${PG_PASSWORD} sslmode=disable" up

migrations-down:
	cd .\db\migrations\ && goose postgres "host=${PG_HOST} port=${PG_PORT} user=${PG_LOGIN} database=${PG_NAME} password=${PG_PASSWORD} sslmode=disable" down

migrations-status:
	cd .\db\migrations\ && goose postgres "host=${PG_HOST} port=${PG_PORT} user=${PG_LOGIN} database=${PG_NAME} password=${PG_PASSWORD} sslmode=disable" status