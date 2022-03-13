include .env
export $(shell sed 's/=.*//' .env)

include .override.env
export $(shell sed 's/=.*//' .override.env)

graph:
	go mod graph | modgv | sfdp -Tpng -o graph.png

api-doc:
	swagger generate spec -o ./swagger.json
	swagger serve -F swagger ./swagger.json

up:
	@echo ---Initializing all services---
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

restart:
	docker-compose restart

destroy:
	@echo "=============Cleaning up============="
	docker-compose down -v
	docker-compose rm -f -v -s
build:
	docker-compose build

run:
	go run main.go

seed:
	docker exec r-api-mongo rm -rf ./seeds
	docker cp ./seeds r-api-mongo:./seeds
	docker exec r-api-mongo mongoimport --username ${MONGO_USERNAME} --password ${MONGO_PASSWORD} --authenticationDatabase admin --db ${MONGO_DATABASE} --collection recipes --file ./seeds/recipes.json --jsonArray
	docker exec r-api-mongo rm -rf ./seeds
