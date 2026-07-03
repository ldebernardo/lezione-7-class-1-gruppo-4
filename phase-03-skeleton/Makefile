.PHONY: test stress race up down health

test:
	docker compose run --rm test

stress:
	docker compose run --rm test go test ./entities -run TestInventoryStressKeepsStockSane -count=10

race:
	docker compose run --rm test go test -race ./...

up:
	docker compose up --build warehouse

down:
	docker compose down

health:
	curl -s http://localhost:8081/health
