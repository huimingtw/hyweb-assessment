dev:
	go run main.go

build:
	go build -o bin/server .

swag:
	swag init -g main.go --output docs

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

