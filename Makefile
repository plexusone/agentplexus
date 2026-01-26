.PHONY: build frontend backend clean dev

build: frontend backend

frontend:
	cd frontend && npm install && npm run build

backend:
	go build -o bin/specui ./cmd/specui

clean:
	rm -rf bin/ internal/server/frontend/*
	touch internal/server/frontend/.gitkeep

dev:
	@echo "Start frontend: cd frontend && npm run dev"
	@echo "Start backend:  go run ./cmd/specui --workspace=$$HOME/go/src/github.com/agentplexus"
