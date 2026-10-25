# Makefile du projet P_Go.
# Sur Windows, `make` n'est pas installé : utilise scripts\dev.ps1 à la place.

GO      ?= go
COMPOSE ?= docker compose

.PHONY: help
help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Met à jour go.mod / go.sum
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Formate le code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Analyse statique du code
	$(GO) vet ./...

.PHONY: build
build: ## Compile les trois binaires dans bin\
	$(GO) build -o bin/pgo-server.exe ./cmd/server
	$(GO) build -o bin/pgo-client.exe ./cmd/client
	$(GO) build -o bin/pgo-admin.exe ./cmd/admin

.PHONY: test
test: ## Lance les tests
	$(GO) test ./... -count=1

.PHONY: run-server
run-server: ## Lance le serveur HTTP
	$(GO) run ./cmd/server

.PHONY: run-client
run-client: ## Lance l'application client (TUI)
	$(GO) run ./cmd/client

.PHONY: run-admin
run-admin: ## Lance l'application administrateur (TUI)
	$(GO) run ./cmd/admin

.PHONY: docker-up
docker-up: ## Démarre PostgreSQL et l'application en conteneurs
	$(COMPOSE) up -d --build

.PHONY: docker-down
docker-down: ## Arrête et supprime les conteneurs
	$(COMPOSE) down

.PHONY: docker-logs
docker-logs: ## Affiche les journaux des conteneurs
	$(COMPOSE) logs -f

.PHONY: psql
psql: ## Ouvre un shell psql sur la base de développement
	$(COMPOSE) exec postgres psql -U pgo -d pgo

.PHONY: check
check: fmt vet test ## Formatage + analyse + tests
