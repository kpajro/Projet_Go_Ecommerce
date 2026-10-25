# P_Go — plateforme d'achat e-commerce en ligne de commande

Projet Go : une plateforme e-commerce dont l'interface est **entièrement en
ligne de commande**. Deux applications (client et administrateur) communiquent
avec un serveur HTTP, lui-même adossé à PostgreSQL via `database/sql`.

Contraintes du sujet respectées :

- **Aucun framework web** : uniquement `net/http`.
- **Aucune ORM** : uniquement `database/sql`, SQL écrit à la main.
- **Charmbracelet** pour l'interface : `bubbletea`, `bubbles`, `huh`, `lipgloss`.
- **Client SSH** pour accéder aux applications à distance.

> Le plan de projet détaillé (calendrier, répartition du travail, API, modèle de
> données) se trouve dans [`PLAN.md`](./PLAN.md).

---

## Prérequis

| Outil | Version | Rôle |
|-------|---------|------|
| [Go](https://go.dev/dl/) | ≥ 1.24 | langage |
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | récent | PostgreSQL 16 en conteneur |
| Git | récent | versionnement |

## Installation

```powershell
git clone <url-du-depot>
cd P_Go
go mod download
```

## Démarrage rapide

```powershell
# 1. Démarrer PostgreSQL (crée la base pgo et la base de test pgo_test)
docker compose up -d postgres

# 2. Lancer le serveur HTTP (applique les migrations au démarrage)
go run ./cmd/server

# 3. Dans un autre terminal : lancer l'interface
go run ./cmd/client
```

Vérification rapide :

```powershell
curl http://localhost:8080/healthz
# {"status":"ok","service":"pgo-server"}
```

Sur Linux/macOS (ou avec `make` installé) les mêmes opérations sont :

```bash
make docker-up
make run-server
make run-client
```

Sur Windows, `make` n'existe pas : utilise `scripts\dev.ps1` (mêmes cibles).

```powershell
.\scripts\dev.ps1 docker-up
.\scripts\dev.ps1 run-server
.\scripts\dev.ps1 check
```

## Configuration

Tout se règle par variables d'environnement (valeurs par défaut entre
parenthèses) :

| Variable | Défaut | Rôle |
|----------|--------|------|
| `PGO_HTTP_ADDR` | `:8080` | adresse d'écoute du serveur |
| `PGO_DATABASE_URL` | `postgres://pgo:pgo@localhost:5433/pgo?sslmode=disable` | chaîne de connexion PostgreSQL |
| `PGO_API_URL` | `http://localhost:8080` | adresse de l'API utilisée par les deux applications CLI |
| `PGO_ENV` | `dev` | contexte d'exécution (`dev`, `test`, `docker`) |

## Structure du dépôt

```
cmd/
  server/      serveur HTTP (net/http) + migrations
  client/      application en ligne de commande pour le client
  admin/       application en ligne de commande pour l'administrateur
internal/
  config/      configuration par variables d'environnement
  store/       couche de persistance (database/sql, SQL écrit à la main)
  tui/         interface Charmbracelet (theme, widgets, écrans)
docs/          architecture, API, guide de soutenance
migrations/    scripts SQL versionnés
```

## Règles de contribution

- `gofmt` et `go vet` doivent passer (`.\scripts\dev.ps1 check`).
- Une fonctionnalité = une branche = une pull request.
- Commits conventionnels : `feat(cart): …`, `fix(auth): …`.
