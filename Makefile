.PHONY: run backend frontend test build generate

run:
	cd frontend && npm run build
	cd backend && GOTOOLCHAIN=go1.25.0 go run ./cmd/server

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

test:
	cd backend && GOTOOLCHAIN=go1.25.0 go test ./...
	cd frontend && npm test
	cd frontend && npm run lint

build:
	cd backend && GOTOOLCHAIN=go1.25.0 go build ./cmd/server
	cd frontend && npm run build

generate:
	cd backend && GOTOOLCHAIN=go1.25.0 go generate ./ent
