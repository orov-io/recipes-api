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
