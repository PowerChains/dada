.PHONY: all setup backend frontend docker-build docker-up dev-backend dev-frontend clean

all: backend frontend

setup:
	cd backend && go mod tidy
	cd frontend && npm install

backend:
	cd backend && CGO_ENABLED=0 go build -o ../bin/dada .

frontend:
	cd frontend && npm install && npm run build

docker-build:
	docker compose build

docker-up:
	docker compose up --build

dev-backend:
	cd backend && go run .

dev-frontend:
	cd frontend && npm install && npm run dev -- --host

clean:
	rm -rf bin frontend/node_modules frontend/dist backend/bin
