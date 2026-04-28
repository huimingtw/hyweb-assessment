dev:
	go run main.go

build:
	go build -o bin/server .

swag:
	swag init --parseDependency --parseInternal -g main.go --output docs

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

