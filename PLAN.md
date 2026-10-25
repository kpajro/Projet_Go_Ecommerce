# P_Go — Plan de projet

> Sujet : plateforme d'achat e-commerce en ligne de commande (Go)
> Barème : 20 points. Contraintes : `net/http` sans framework, `database/sql` sans framework, Charmbracelet, client SSH.
> Livrable : **lien Git** + dépôt sur MyGES.

---

## 1. Décryptage du sujet

### 1.1 Périmètre fonctionnel (points)

| # | Fonctionnalité | Points | Contrainte forte |
|---|----------------|--------|------------------|
| F1 | Authentification (inscription, confirmation par code, login, reset mot de passe) | 1 | — |
| F2 | Recherche produit (nom, prix, description, catégorie, prix TTC) | 2 | — |
| F3 | Panier (ajout, quantités, total TTC, lié à un compte) | 2 | panier **persisté** en BDD |
| F4 | Paiement (carte, expiration, CVC — **sans Stripe**) | 2 | paiement simulé |
| F5 | Commandes (historique, statuts, date, ID métier `CMD-…`) | 2 | 5 statuts dont *annulé + raison* |
| F6 | Gestion commandes admin (changement de statut, création pour un client) | 1 | — |
| F7 | Gestion utilisateurs admin (CRUD + confirmation) | 1 | — |
| F8 | Gestion produits admin (création, ID métier `PDT-…`) | 1 | — |
| F9 | Interfaces | 2 | **2 applications CLI séparées** : client + admin |
| F10 | Persistance | 1 | `database/sql`, PostgreSQL (Docker) **ou** SQLite |
| F11 | Charmbracelet | 3 | bubbletea + bubbles + huh + lipgloss, UI « vivante » |
| F12 | Interface SSH | 2 | client SSH pour piloter les 2 applications |
| | **Total** | **20** | |

### 1.2 Contraintes à ne jamais violer (pénalités)

- ❌ **Aucun framework web** → uniquement `net/http` (le `ServeMux` de Go 1.22+ avec patterns `GET /api/v1/products/{id}` suffit).
- ❌ **Aucun ORM / driver hors `database/sql`** → SQL écrit à la main, requêtes préparées.
- ✅ **Uniquement Charmbracelet** pour l'UI (`bubbletea`, `bubbles`, `huh`, `lipgloss`).
- ⚠️ `golang.org/x/crypto` (bcrypt, ssh) et un driver DB ne sont **pas** des frameworks : ce sont des modules std-adjacents. À justifier en soutenance.
- ⚠️ **Code maîtrisé** : tu dois pouvoir expliquer chaque fichier, schéma SQL et requête. Pas de code collé sans compréhension.
- ⚠️ **Livraison** : dépôt Git accessible + dépôt du livrable sur MyGES + lien Git (pas de ZIP en guise de livrable).
- ⚠️ **Temps** : prévoir une marge de 20 % avant la date limite (soutenance = stress + bugs de dernière minute).

### 1.3 Identifiants métier

Format : **PREFIXE-6 CARACTÈRES** (ex. `BSK-1KH8E7`, `CMD-1F2S8B`, `PDT-7D2K8N`).

- Préfixes : `BSK` panier, `CMD` commande, `PDT` produit.
- Alphabet sans caractères ambigus : `23456789ABCDEFGHJKLMNPQRSTUVWXYZ` (exclut 0/O/1/I).
- Génération via `crypto/rand` (jamais `math/rand`), collision ⇒ retry sur violation d'unicité.

---

## 2. Architecture cible

```
┌──────────────┐        HTTP/JSON (net/http)        ┌──────────────────────────┐
│ pgo-client   │ ─────────────────────────────────▶  │  pgo-server              │
│ (bubbletea)  │                                     │   net/http (ServeMux)    │
├──────────────┤                                     │   handlers + middleware  │
│ pgo-admin    │ ─────────────────────────────────▶  │   services (logique)     │
│ (bubbletea)  │                                     │   store (database/sql)   │
└──────────────┘                                     └───────────┬──────────────┘
      ▲                                                              │
      │  x/crypto/ssh (client SSH)                                  ▼
┌─────┴──────────────────────────────────────────────────┐  ┌───────────────────┐
│ pgo-ssh : connect --host srv --mode client|admin        │  │ PostgreSQL        │
│ (pty + exec de pgo-client / pgo-admin à distance)       │  │ (conteneur Docker)│
└────────────────────────────────────────────────────────┘  └───────────────────┘
```

### 2.1 Arborescence

```
P_Go/
├── go.mod                        module github.com/<ton-pseudo>/pgo
├── cmd/
│   ├── server/main.go            API + (option) serveur SSH embarqué
│   ├── client/main.go            TUI utilisateur
│   ├── admin/main.go             TUI administrateur
│   └── sshcli/main.go            client SSH (lance la TUI à distance)
├── internal/
│   ├── domain/                   structs métier + statuts + règles de transition
│   ├── bizid/                    génération des IDs métier (BSK/CMD/PDT)
│   ├── security/                 bcrypt, token HMAC, validation carte (Luhn)
│   ├── store/                    interface Store + implémentation PostgreSQL
│   │   └── postgres/             SQL PostgreSQL (pgx stdlib)
│   ├── service/                  auth, catalog, cart, order, payment, admin
│   ├── httpapi/                  router, handlers, middleware, json, errors
│   ├── apiclient/                client HTTP partagé par client & admin
│   └── tui/
│       ├── theme/                styles lipgloss
│       ├── widgets/              composants bubbles partagés
│       ├── client/               modèles bubbletea (screens client)
│       └── admin/                modèles bubbletea (screens admin)
├── migrations/                   *.sql versionnés
├── scripts/seed.go                jeu de données de démo
├── docker/                       Dockerfile multi-stage
├── docker-compose.yml             postgres (bases pgo + pgo_test) + app
├── Makefile
├── README.md
├── docs/                         architecture.md, api.md, soutenance.md
└── .github/workflows/ci.yml       go vet + go test
```

> **Règle de dépendance** : `domain` ← `service` ← `httpapi` ← `cmd/*`. `tui` ne connaît que `apiclient` (jamais la BDD). Le serveur ne connaît jamais bubbletea.

### 2.2 Choix techniques

| Sujet | Choix | Justification |
|-------|-------|---------------|
| HTTP | `net/http` + `ServeMux` patterns Go 1.22+ | pas de framework, routing lisible |
| BDD | **PostgreSQL 16 uniquement**, en conteneur Docker | l'énoncé propose PG *ou* SQLite : on prend PG (1er cité) et on n'implémente qu'un seul driver, donc un seul `Store` à maintenir et à expliquer |
| BDD tests | base `pgo_test` sur le même conteneur, **un schéma éphémère par exécution** (`test_<rand>`, `DROP SCHEMA` en fin de test) | pas de 2e driver, pas de CGO, tests isolés et reproductibles |
| Driver PG | `github.com/jackc/pgx/v5/stdlib` | implémente `database/sql` |
| Mots de passe | `golang.org/x/crypto/bcrypt` | standard, cost 12 |
| Session | Token HMAC-SHA256 maison (stdlib : `crypto/hmac`, `crypto/sha256`, `base64`) | pas de dépendance JWT, faciles à expliquer |
| Prix | `price_htc` stocké (NUMERIC 10,2), **TTC calculé** à 20 % | évite la dérive et respecte « prix TTC » de l'énoncé |
| Carte bancaire | validation Luhn + expiration + CVC, **jamais stockée** (4 derniers chiffres seulement) | bonne pratique, valorise la soutenance |
| Identifiants | `crypto/rand` + alphabet 32 sans ambiguïté | conforme aux exemples `BSK-1KH8E7` |
| Logs | `log/slog` (stdlib) | structured logs serveur |
| pty serveur SSH | `github.com/creack/pty` (serveur Linux) | micro-lib système (pas un framework) pour un vrai terminal côté SSH |
| Erreurs API | JSON `{"error":{"code":"...","message":"..."}}` | contrat clair entre CLI et API |

---

## 3. Modèle de données (PostgreSQL)

> **Un seul driver, un seul `Store`.** Le schéma est écrit en SQL portable (`NUMERIC`, `TEXT`, `TIMESTAMPTZ`) : aucune table n'est liée à un type propriétaire d'un autre SGBD. Les requêtes sont écrites à la main dans `internal/store/postgres/*.go` (jamais d'ORM), avec requêtes préparées et `context.Context` partout.

```sql
CREATE TABLE users (
  id                BIGSERIAL PRIMARY KEY,
  email             TEXT UNIQUE NOT NULL,          -- stocké en minuscules
  password_hash     TEXT NOT NULL,
  role              TEXT NOT NULL DEFAULT 'client' CHECK (role IN ('client','admin')),
  is_confirmed     BOOLEAN NOT NULL DEFAULT FALSE,
  confirmation_code TEXT UNIQUE,
  reset_code        TEXT,
  reset_expires_at  TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
  id           BIGSERIAL PRIMARY KEY,
  business_id  TEXT UNIQUE NOT NULL,               -- PDT-7D2K8N
  name         TEXT NOT NULL,
  description  TEXT NOT NULL DEFAULT '',
  category     TEXT NOT NULL DEFAULT 'général',
  price_htc    NUMERIC(10,2) NOT NULL CHECK (price_htc >= 0),
  stock        INT NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_name     ON products (lower(name));
CREATE INDEX idx_products_category ON products (lower(category));
CREATE INDEX idx_products_price    ON products (price_htc);

CREATE TABLE carts (
  id          BIGSERIAL PRIMARY KEY,
  business_id TEXT UNIQUE NOT NULL,                -- BSK-1KH8E7
  user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','converted','abandoned')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_carts_active_user ON carts (user_id) WHERE status = 'active';

CREATE TABLE cart_items (
  id              BIGSERIAL PRIMARY KEY,
  cart_id         BIGINT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
  product_id      BIGINT NOT NULL REFERENCES products(id),
  quantity        INT NOT NULL CHECK (quantity > 0),
  unit_price_htc  NUMERIC(10,2) NOT NULL,
  UNIQUE (cart_id, product_id)
);

CREATE TABLE orders (
  id                 BIGSERIAL PRIMARY KEY,
  business_id        TEXT UNIQUE NOT NULL,        -- CMD-1F2S8B
  user_id            BIGINT NOT NULL REFERENCES users(id),
  status             TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending','paid','shipped','delivered','cancelled')),
  cancellation_reason TEXT,
  total_htc          NUMERIC(10,2) NOT NULL DEFAULT 0,
  total_ttc          NUMERIC(10,2) NOT NULL DEFAULT 0,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user ON orders (user_id, created_at DESC);

CREATE TABLE order_items (
  id            BIGSERIAL PRIMARY KEY,
  order_id      BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id    BIGINT REFERENCES products(id) ON DELETE SET NULL,
  business_id   TEXT NOT NULL,                     -- snapshot PDT-…
  product_name  TEXT NOT NULL,                     -- snapshot
  quantity      INT NOT NULL CHECK (quantity > 0),
  unit_price_htc NUMERIC(10,2) NOT NULL
);

CREATE TABLE payments (
  id          BIGSERIAL PRIMARY KEY,
  order_id    BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  card_brand  TEXT,                -- Visa, Mastercard…
  card_last4  TEXT NOT NULL,       -- JAMAIS le numéro complet
  expiry      TEXT NOT NULL,
  amount_ttc  NUMERIC(10,2) NOT NULL,
  status      TEXT NOT NULL DEFAULT 'captured' CHECK (status IN ('captured','refused')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Machine à statuts de commande** (à implémenter et tester) :

```
pending ──pay──▶ paid ──ship──▶ shipped ──deliver──▶ delivered
   │                 │
   └──cancel──▶ cancelled (raison obligatoire)
```

**Stratégie de tests d'intégration sans 2e SGBD** : les tests tournent sur la base `pgo_test` du même conteneur PostgreSQL. Le helper `internal/testutil` crée un schéma au nom aléatoire (`test_<rand>`), y applique les migrations, positionne `search_path` sur ce schéma pour la connexion de test, puis le supprime en `t.Cleanup`. Chaque test repart donc d'une base propre, sans jamais toucher la base de dev.

---

## 4. API REST (contrat entre le serveur et les 2 CLI)

Base : `/api/v1`. Auth : `Authorization: Bearer <token>`.

### Auth
| Méthode | Route | Description |
|---|---|---|
| POST | `/auth/register` | `{email,password}` → crée le compte, génère le code de confirmation |
| POST | `/auth/confirm` | `{email,code}` → confirme le compte |
| POST | `/auth/login` | `{email,password}` → `{token,user}` (refusé si non confirmé) |
| POST | `/auth/password/forgot` | `{email}` → envoie un code de reset (toujours 200, pas d'énumération d'utilisateur) |
| POST | `/auth/password/reset` | `{email,code,new_password}` → réinitialise |

### Catalogue (lecture publique, écriture admin)
| Méthode | Route | Description |
|---|---|---|
| GET | `/products` | filtres `?q=&name=&description=&category=&price_min=&price_max=&ttc_min=&ttc_max=&sort=&page=&limit=` |
| GET | `/products/{id}` | détail (accepte l'ID métier `PDT-…`) |
| GET | `/categories` | liste des catégories (autocomplétion) |

### Panier (utilisateur connecté)
| Méthode | Route | Description |
|---|---|---|
| GET | `/cart` | panier actif + total TTC |
| POST | `/cart/items` | `{product_id, quantity}` (ou tableau pour l'ajout multiple) |
| PATCH | `/cart/items/{productId}` | `{quantity}` (0 = suppression) |
| DELETE | `/cart/items/{productId}` | suppression de la ligne |

### Commandes & paiement (utilisateur connecté)
| Méthode | Route | Description |
|---|---|---|
| POST | `/orders` | crée la commande depuis le panier → `CMD-…`, panier vidé, **transaction SQL** |
| GET | `/orders` | mes commandes + statut |
| GET | `/orders/{id}` | détail (lignes, totaux, paiement) |
| POST | `/orders/{id}/pay` | `{card_number,expiry,cvc}` → validation + passage à `paid` |
| POST | `/orders/{id}/cancel` | `{reason}` → `cancelled` |

### Admin (rôle `admin` obligatoire)
| Méthode | Route | Description |
|---|---|---|
| GET | `/admin/orders` | toutes les commandes |
| PATCH | `/admin/orders/{id}/status` | `{status, reason?}` |
| POST | `/admin/orders` | `{user_id, items:[{product_id,quantity}]}` |
| GET/POST | `/admin/users` | lister / créer |
| PATCH/DELETE | `/admin/users/{id}` | modifier / supprimer |
| POST | `/admin/users/{id}/confirm` | confirmer un compte |
| GET/POST | `/admin/products` | lister / créer (ID `PDT-…` auto) |
| PATCH/DELETE | `/admin/products/{id}` | modifier / supprimer |

### Codes HTTP & erreurs
`200 / 201 / 204` · `400` (validation) · `401` (non authentifié) · `403` (droits) · `404` · `409` (conflit, ex. email déjà pris) · `422` (règle métier, ex. transition de statut invalide) · `500`.
Corps d'erreur unique : `{"error":{"code":"CART_EMPTY","message":"…"}}`.

---

## 5. Les deux interfaces CLI (Charmbracelet — 3 points)

### 5.0 Qui fait quoi dans l'écosystème Charmbracelet

L'énoncé cite **4 composants** : les utiliser tous est ce qui rapporte les 3 points. Un `bubbletea` seul avec des `fmt.Println` ferait perdre la moitié de la note.

| Librairie | Rôle | Où l'utiliser dans P_Go |
|---|---|---|
| **`bubbletea`** | Le moteur : architecture `Model` / `Update` / `View` + `tea.Cmd` (asynchrone) | Squelette de **toutes** les applications. Un modèle racine par app, un modèle par écran, machine à états d'écrans |
| **`bubbles`** | Composants prêts à l'emploi, construits **sur** bubbletea | `textinput` (recherche, code de confirmation), `table` (produits, panier, commandes, users), `viewport` (détail produit), `spinner` (chargement), `key` (raccourcis), `help` (barre d'aide) |
| **`huh`** | Formulaires : `huh.NewInput`, `huh.NewSelect`, `huh.NewConfirm` — **basés sur bubbletea** | Login, inscription, mot de passe oublié, confirmation de code, paiement (carte/expiration/CVC), création de produit, création de commande admin, suppression avec confirmation |
| **`lipgloss`** | Rendu pur : styles, couleurs, bordures, mise en page (aucune logique) | Thème centralisé `internal/tui/theme` : palette, `lipgloss.NewStyle()`, badge de statut, header/footer, largeur adaptative |

**Règle d'or** : `lipgloss` ne décide rien, `bubbles` gère l'interaction d'un widget, `huh` gère un formulaire complet, `bubbletea` orchestre. Un seul package connaît `lipgloss` (le thème) → changer de couleur = 1 fichier.

**Pourquoi c'est « réactif et vivante »** (le vocabulaire de l'énoncé) : chaque appel HTTP part en `tea.Cmd` (goroutine) et revient sous forme de message ; `Update` ne bloque jamais l'affichage ; un `spinner` tourne pendant l'appel ; `tea.Batch` lance plusieurs requêtes en parallèle. L'UI reste instantanée même si l'API met 2 secondes.

### 5.1 Principes TUI

- Un **modèle bubbletea unique par application** (`client.AppModel`, `admin.AppModel`) avec un état `screen` + un modèle par écran.
- Toute appel HTTP = `tea.Cmd` qui retourne un message (`loginOKMsg`, `productsLoadedMsg`, `apiErrMsg`…). **Jamais d'appel HTTP bloquant dans `Update`**.
- `bubbles/spinner` pendant les chargements, `bubbles/textinput` pour la recherche, `bubbles/table` pour les listes, `bubbles/viewport` pour le détail, `huh` pour les formulaires.
- `lipgloss` : thème centralisé (couleurs ANSI, bordures, largeur adaptative), aucune couleur codée en dur dans les écrans.
- `tea.Quit` / `tea.KeyMsg` : `q` pour quitter, `?` pour l'aide, `ctrl+c` pour annuler.

### 5.2 Écrans — application client

| Écran | Composants | Actions |
|---|---|---|
| 1. Connexion / Inscription | `huh` form + onglets | login, register, forgot password |
| 2. Confirmation de compte | `huh` (code) | saisie du code à 6 caractères |
| 3. Catalogue | `textinput` recherche + `table` + filtres | recherche multi-critères, tri par prix/TTC, sélection |
| 4. Détail produit | `viewport` + actions | description, prix HT/TTC, ajout au panier (quantité) |
| 5. Panier | `table` éditable + footer | +/− quantité, suppression, total TTC, bouton « payer » |
| 6. Paiement | `huh` (carte, expiration, CVC) | validation, retour d'erreur, confirmation |
| 7. Mes commandes | `table` + badges de statut | historique, détail, suivi, annulation |

### 5.3 Écrans — application admin

| Écran | Actions |
|---|---|
| 1. Connexion admin | login + vérification du rôle |
| 2. Tableau de bord | menu de navigation + compteurs (nb commandes, CA, nb users) |
| 3. Commandes | table, changer statut, créer une commande pour un client, annuler avec raison |
| 4. Utilisateurs | table, créer/modifier/supprimer/confirmer un compte |
| 5. Produits | table, créer/modifier/supprimer un produit (ID `PDT-…` auto) |

### 5.4 Détails qui font la différence à la soutenance

- Spinner + désactivation des actions pendant le chargement (donc « réactive et vivante »).
- Notifications temporisées (toast) sur succès/erreur.
- `huh` avec validation intégrée (email, mot de passe, CVC).
- Code couleur par statut de commande (pending/paid/shipped/delivered/cancelled).
- Confirmation par `huh.Confirm` avant toute suppression.

---

## 6. SSH (2 points)

**Deux briques complémentaires :**

1. **Client SSH (exigé par l'énoncé)** — `cmd/sshcli` :
   - `golang.org/x/crypto/ssh` : dial, authentification (mot de passe ou clé), `RequestPty("xterm-256color", h, w)`, `session.Run("pgo-client")`.
   - StdIn/StdOut branchés sur la session, `SIGWINCH` → `session.WindowChange` (taille de fenêtre réelle).
   - CLI : `pgo-ssh connect --host srv --port 22 --user admin --mode client|admin`.
2. **Serveur SSH embarqué (pour la démo)** — `cmd/server --ssh-listen :2222` :
   - Clé hôte ED25519 générée et persistée (`ssh/host_key`), auth password contre la table `users`, rôle lu en base.
   - À la connexion : fork de `pgo-client` ou `pgo-admin` avec stdin/stdout = channel SSH (le client demande un pty ⇒ le serveur tourne dans un vrai terminal ⇒ bubbletea en raw mode fonctionne).
   - Conséquence : un simple `ssh -p 2222 admin@localhost` donne la TUI complète. Démo très convaincante.

> Le serveur tourne en conteneur Linux (Docker) : c'est là que les pty fonctionnent. Le poste de dev Windows ne sert que de client.

---

## 7. Plan de travail (10 semaines + marge)

> Adapte les durées : 8-10 h de développement effectif par semaine. Chaque phase se termine par un **commit taggé** (`v0.1-auth`, …) ⇒ historique Git propre pour la soutenance.

### Phase 0 — Mise en route (½ journée)
- [ ] Installer **Go 1.24+** (`go version` OK) — absent de ta machine actuellement.
- [ ] `git init`, `.gitignore` (binaire, `.env`, `data/`), commit initial.
- [ ] `go mod init`, arborescence des dossiers vide.
- [ ] `docker compose up -d postgres` + vérif que le port est joignable.
- [ ] Spikes de risque (2 h) : `bubbletea` hello world, `pgx` + `database/sql` query, `x/crypto/ssh` dial vers un conteneur `linuxserver/openssh-server`.

### Phase 1 — Socle serveur + BDD (semaine 1) — *F10*
- [ ] `migrations/001_init.sql` (schéma ci-dessus) + runner de migrations (`internal/store/migrate.go`, table `schema_migrations`).
- [ ] `internal/store` : interface `Store` + implémentation PostgreSQL.
- [ ] `internal/httpapi` : `ServeMux`, middlewares (JSON, logs, CORS non nécessaire, auth, rôle admin), helper JSON.
- [ ] `GET /healthz` + `GET /api/v1/version`.
- [ ] `Makefile` : `make migrate`, `make run-server`, `make seed`, `make test`.
- [ ] Dockerfile multi-stage + `docker-compose.yml` (postgres + app) ; le conteneur PG crée **deux bases** : `pgo` (dev) et `pgo_test` (tests).
- [ ] `internal/testutil` : helper de tests d'intégration — crée un **schéma éphémère** `test_<rand>` (migrations appliquées dedans, `DROP SCHEMA CASCADE` en `t.Cleanup`), pour que les tests soient isolés et rejouables.
- ✅ *Critère de fin* : `curl localhost:8080/healthz` répond, la BDD est créée, Docker up fonctionne, `make test` passe sur `pgo_test`.

### Phase 2 — Authentification (semaine 2) — *F1*
- [ ] `internal/security` : bcrypt, token HMAC, générateur de code aléatoire.
- [ ] Service `auth` : register (user non confirmé), confirm, login (refus si non confirmé), forgot (génère un reset code, journalise), reset.
- [ ] Middleware `RequireAuth` / `RequireAdmin` (extraction du token, vérification signature + expiration).
- [ ] Handlers + routes + codes HTTP (409 email déjà pris, 401...).
- [ ] Tests unitaires : hash ≠ mot de passe, token signature/expiration, transition de confirmation.
- ✅ *Critère de fin* : inscription → confirmation → login → appel protégé ; reset fonctionne.

### Phase 3 — Catalogue & recherche (semaine 3) — *F2, F8 (partiel)*
- [ ] `internal/bizid` : générateur `PDT-…` + test de format et d'unicité.
- [ ] Service `catalog` : list avec filtres (nom, description, catégorie, prix, TTC), tri, pagination.
- [ ] Handler `GET /products` + `GET /products/{id}` (accepte `PDT-…` ou l'id numérique).
- [ ] `scripts/seed.go` : 30-50 produits, 4-5 catégories, 2 comptes de démo (client + admin) — indispensable pour la démo.
- [ ] Test d'intégration : filtre par catégorie, par plage de prix, tri par TTC, pagination.
- ✅ *Critère de fin* : la recherche combine plusieurs critères et renvoie des totaux TTC justes.

### Phase 4 — Panier (semaine 4) — *F3*
- [ ] Panier créé à la volée au premier ajout, lié à l'utilisateur, ID `BSK-…`.
- [ ] Ajout (unitaire + multiple), modification de quantité (bornée 1..99), suppression.
- [ ] Total HT / TVA / **TTC** calculés côté service (TVA 20 %, livraison gratuite).
- [ ] Handlers + routes ; le panier survit au redémarrage du serveur (BDD).
- [ ] Tests : total TTC exact, fusion des lignes de même produit, suppression de ligne.
- ✅ *Critère de fin* : le panier persiste et affiche un total TTC correct.

### Phase 5 — Commande & paiement (semaine 5) — *F4, F5*
- [ ] Transaction SQL : conversion panier → commande (`CMD-…`), vidage du panier, snapshot des prix/noms.
- [ ] Machine à statuts + transitions autorisées (422 sinon).
- [ ] `internal/security` : validation carte (Luhn, expiration future, CVC 3/4 chiffres, détection de marque).
- [ ] Paiement simulé : `4242 4242 4242 4242` = accepté, `4000 0000 0000 0002` = refusé, `4000 0000 0000 9995` = fonds insuffisants (scénarios de démo).
- [ ] Stockage limité à `card_last4` + marque + expiration.
- [ ] `GET /orders`, `GET /orders/{id}`, `/pay`, `/cancel` (raison obligatoire).
- [ ] Tests : total commande = total panier, refus sur carte invalide, transitions interdites.
- ✅ *Critère de fin* : on paie un panier sauvegardé, on retrouve la commande avec son statut.

### Phase 6 — API Admin (semaine 6) — *F6, F7, F8*
- [ ] CRUD utilisateurs (create/update/delete/confirm) + garde `RequireAdmin`.
- [ ] CRUD produits (create avec ID `PDT-…` auto, update, delete) + recherche admin.
- [ ] Gestion commandes : liste globale, changement de statut avec raison, création de commande pour n'importe quel client (multi-produits).
- [ ] Tests : un client ne peut pas taper sur `/admin/*` (403).
- ✅ *Critère de fin* : toutes les actions admin sont possibles via l'API.

### Phase 7 — TUI Client (semaines 7-8) — *F9, F11*
- [ ] Socle : `theme` lipgloss, `apiClient` (client HTTP + timeout + erreurs typées), modèle racine bubbletea + machine à états d'écrans.
- [ ] Écran Connexion/Inscription (huh) + mot de passe oublié + confirmation de code.
- [ ] Écran Catalogue : recherche live (débounce 300 ms), filtres, tri, table, sélection.
- [ ] Écran Détail produit + ajout au panier (quantité).
- [ ] Écran Panier : édition des quantités, total TTC animé, accès au paiement.
- [ ] Écran Paiement (huh) + écran Commandes (table + badges + détail).
- [ ] Polish : spinners, toasts, raccourcis, aide `?`, gestion d'erreur réseau (serveur injoignable).
- ✅ *Critère de fin* : parcours complet utilisable au clavier, sans crash.

### Phase 8 — TUI Admin (semaine 9) — *F9, F11*
- [ ] Même socle, modèlesbubbletea dédiés.
- [ ] Écrans Commandes / Utilisateurs / Produits / Dashboard.
- [ ] Formulaires huh de création/modification, confirmations de suppression.
- [ ] ✅ *Critère de fin* : l'admin fait tout depuis sa TUI, y compris changer un statut et créer une commande pour un client.

### Phase 9 — SSH (semaine 9-10) — *F12*
- [ ] `cmd/sshcli` : dial, pty, exec, resize, erreurs (auth, timeout).
- [ ] Serveur SSH embarqué dans `cmd/server` : clé hôte, auth par la BDD, fork de la TUI, mapping rôle → application.
- [ ] Test réel : `pgo-ssh connect --host ... --mode admin` ⇒ TUI admin à distance, resize correct.
- [ ] Documentation dans le README + capture pour la soutenance.
- ✅ *Critère de fin* : la démo SSH fonctionne en direct.

### Phase 10 — Qualité, docs & soutenance (semaine 10-11)
- [ ] Tests : `go test ./...` (unitaires services avec un faux `Store` + intégration HTTP sur PostgreSQL `pgo_test` via schéma éphémère), couverture ≥ 60 % sur `service` et `httpapi`.
- [ ] Qualité : `gofmt`, `go vet`, `golangci-lint` (optionnel), pas de `panic` non géré, erreurs contextuelles.
- [ ] `README.md` : installation, lancement, docker compose, comptes de démo, captures d'écran, architecture.
- [ ] `docs/architecture.md` (schémas + justification des choix), `docs/api.md`, `docs/db.md`.
- [ ] `docs/soutenance.md` : plan de slides + script de démo + 15 questions probables + réponses.
- [ ] Dépôt Git propre : commits conventionnels, `README` soigné, `.gitignore`, licence, `Makefile` lisible.
- [ ] Repo GitHub/Forge accessible **en lecture** (pas de fork) + upload du livrable sur **MyGES** + le **lien Git** dans le message de rendu.
- [ ] 2 répétitions de soutenance chrono (10-15 min) + 1 répétition « sans réseau » (captures/vidéos de secours).
- ✅ *Critère de fin* : démo complète en < 10 min, questions sans trou.

### Version express (4 semaines, si la date limite est proche)
Sem. 1 : Phases 0+1+2 · Sem. 2 : Phases 3+4 · Sem. 3 : Phases 5+6 · Sem. 4 : Phases 7 (TUI client minimum viable) + Phase 9 (SSH minimum) + doc/soutenance.
→ Priorité aux 14 points les plus accessibles : F1, F2, F3, F4, F5, F9, F10, F11, F12. L'admin (F6/F7/F8) est maintenu mais allégé.

---

## 7.1 Planning daté (21 octobre → 18 décembre 2026)

> 8 semaines. **Règle d'or : gel des fonctionnalités le 11 décembre.** La dernière semaine ne sert qu'à la doc, aux tests et aux répétitions — jamais à du code nouveau.

| Semaine | Dates | Phase | Focalisation | Points visés |
|---|---|---|---|---|
| S0 | mer. 21 → dim. 25 oct | **Phase 0** Mise en route | Go installé, git, docker, spikes (bubbletea / pgx / ssh) | — |
| S1 | lun. 26 oct → dim. 1 nov | **Phase 1** Socle BDD + serveur | schéma SQL, migrations, store, `ServeMux`, Dockerfile, `Makefile`, **contrat d'API figé** | F10 |
| S2 | lun. 2 → dim. 8 nov | **Phase 2** Auth | bcrypt, token HMAC, register/confirm/login/reset, middleware rôles | F1 |
| S3 | lun. 9 → dim. 15 nov | **Phase 3** Catalogue | recherche multi-critères, `PDT-…`, seed 40 produits | F2 |
| S4 | lun. 16 → dim. 22 nov | **Phase 4** Panier | panier `BSK-…` persisté, quantités, total TTC | F3 |
| S5 | lun. 23 → dim. 29 nov | **Phase 5** Commande + paiement | transaction panier→commande `CMD-…`, Luhn, statuts | F4, F5 |
| S6 | lun. 30 nov → dim. 6 déc | **Phase 6** API admin + TUI client v1 | CRUD admin, **première TUI client complète** | F6, F7, F8, F11 |
| S7 | lun. 7 → dim. 13 déc | **Phase 9** SSH + TUI admin + tests | client SSH, serveur SSH embarqué, TUI admin, `go test` | F12, F9, F11 |
| S8 | lun. 14 → ven. 18 déc | **Phase 10** Doc + soutenance | README, docs, script de démo, répétitions, **soutenance ven. 18** | — |

**⚠️ Vacances de Toussaint** (fin oct. → début nov. en général) : la disponibilité est réduite en S0/S1. Décale l'effort vers S2 et anticipates le socle BDD (S1) pendant la semaine de Toussaint — c'est du travail calme, parfait pour ça.

**Jalons de contrôle** (à cocher avec le groupe chaque dimanche) :
- **3 nov** : le serveur répond sur `/healthz` et la BDD Docker tourne → *sinon, on.reduce la portée TUI, jamais la BDD*.
- **15 nov** : recherche + panier OK de bout en bout (même si la TUI est encore console).
- **29 nov** : commande payée + statut visible.
- **6 déc** : TUI client complète et fluide (c'est le plus gros risque du projet).
- **11 déc** : **gel des fonctionnalités** — plus aucune feature, seulement correctifs et doc.
- **16 déc** : 1re répétition de soutenance.
- **18 déc** : soutenance.

---

## 7.2 Répartition des 3 membres (anti-conflits Git)

Le principe : **un module = un propriétaire**. On ne modifie jamais le module d'un autre sans PR.

| | **Membre A — Backend / Données** | **Membre B — TUI Client** | **Membre C — TUI Admin + SSH** |
|---|---|---|---|
| Dossiers | `internal/domain`, `internal/store/*`, `internal/service/*`, `internal/httpapi/*`, `internal/apiclient`, `internal/bizid`, `internal/security` | `internal/tui/theme`, `internal/tui/widgets`, `internal/tui/client/*` | `internal/tui/admin/*`, `internal/sshserver`, `cmd/sshcli` |
| Binaires | `cmd/server` | `cmd/client` | `cmd/admin`, `cmd/sshserver` |
| Livraison | F10, F1, F2, F3, F4, F5, F6, F7, F8 (API) | F9 client, F11 client, UX de F1→F5 | F9 admin, F11 admin, F12, UX de F6→F8 |
| Points d'appui | fournit `apiclient` (contrat HTTP typé) aux 2 autres | fournit `theme` + `widgets` à C (PR) | fournit le serveur SSH à A (PR) |

### Règles de non-conflit
1. **`internal/domain` et `internal/apiclient` appartiennent à A.** B et C n'y touchent pas : ils définissent leurs propres **view-models** dans leur package `tui/...` (découplage propre, et bonne réponse en soutenance sur le MVVM).
2. **`internal/tui/theme` et `internal/tui/widgets` appartiennent à B.** C envoie une PR pour y ajouter un widget.
3. **`cmd/sshserver` est un binaire séparé** de `cmd/server` : C ne touche jamais au code de A, et le serveur SSH se connecte à l'API via `PGO_API_URL` (aucun couplage direct).
4. Le **contrat d'API est figé en fin de S1** (`docs/api.md` = source de vérité). Toute évolution passe par une PR de A.

### Rituel Git (30 min chaque vendredi)
- Branches longues : `dev/a-backend`, `dev/b-tui-client`, `dev/c-tui-admin-ssh`.
- Merge dans `integration` le vendredi, **testé à trois** (`make run-server` + TUI) avant merge dans `main`.
- Un **tag par feature** : `v0.3-cart`, `v0.5-order`… → historique narratif pour la soutenance.
- Commits conventionnels : `feat(cart): ajout de la modification de quantité`.

### Pour la soutenance
Chacun doit **démo sa partie** ET **expliquer l'ensemble** (schéma BDD, flux HTTP, machine à statuts). Prépare une passe de **peer-review** : chaque membre rejoue la démo de l'autre et pose 3 questions. C'est la meilleure protection contre la pénalité « participation inégale ».

---

## 7.3 Détail SSH : le pty côté serveur

Le serveur SSH embarqué doit lancer la TUI **dans un vrai terminal**, sinon bubbletea tombe en mode lecture ligne à ligne (flèches et couleurs dégradées).

- Côté **client** (ce qu'exige l'énoncé) : `x/crypto/ssh` pur — `RequestPty("xterm-256color")`, `session.Run()`, `SIGWINCH` → `session.WindowChange`, `x/term.GetSize` pour la taille réelle. Aucun problème.
- Côté **serveur** : le binaire `cmd/sshserver` (conteneur Linux) utilise `github.com/creack/pty` pour allouer un pty maître/esclave, y lance `pgo-client`/`pgo-admin` selon le rôle lu en base, puis relaie pty ⇄ canal SSH et applique le redimensionnement. `creack/pty` est une micro-bibliothèque système, pas un framework — à justifier.
- Le serveur tourne **sous Linux (Docker)** : c'est là que le pty est naturel. Ta machine Windows ne joue que le rôle de client SSH.

---

## 8. Tableau de bord des points (à tenir à jour)

| Sprint | Points visés | Cumul |
|---|---|---|
| Fin P1 | F10 (1) | 1 |
| Fin P2 | F1 (1) | 2 |
| Fin P3 | F2 (2) + F8 partiel | 4 |
| Fin P4 | F3 (2) | 6 |
| Fin P5 | F4 (2) + F5 (2) | 10 |
| Fin P6 | F6 (1) + F7 (1) + F8 (1) | 13 |
| Fin P7/P8 | F9 (2) + F11 (3) | 18 |
| Fin P9 | F12 (2) | 20 |

---

## 9. Risques & parades

| Risque | Impact | Parade |
|---|---|---|
| `charmbracelet` v2 vs v1, API changeante | bloquant | Épingler les versions dans `go.mod` dès le Phase 0, s'y tenir |
| pty / SSH bancal sous Windows | élevé | Le **serveur** tourne en conteneur Linux ; seul le client tourne sous Windows |
| Un seul driver PG à maintenir | faible | PostgreSQL unique, `Store` interface + un fake pour les tests unitaires |
| Retard sur la TUI (3 points !) | critique | Commencer le socle TUI dès la Phase 2, en parallèle de l'API |
| SQL manual trop long | moyen | Écrire un helper de requête, garder les transactions explicites |
| Enchaînement de phases = perdre | critique | Une seule feature = commit + tag ; ne jamais accumuler 3 jours de code non commité |
| « Utilisation abusive de l'IA » | pénalité | Tu dois pouvoir redessiner chaque écran et chaque requête **sans** l'IA. Rédiger le code à la main, relire, commenter ce qui est subtil |
| Temps de soutenance | pénalité | Script de démo chronométré + plan B hors-ligne (vidéo 90 s) |
| Conflits Git entre les 3 membres | critique | Propriété exclusive des modules, contrat d'API figé en S1, merge ritual vendredi |
| Une personne en avance, deux en retard | critique | Chaque module a un propriétaire **et** un binôme de relecture ; démo croisée obligatoire |

---

## 10. Commandes de démarrage (à exécuter maintenant)

```powershell
# 1. Installer Go 1.24+ : https://go.dev/dl/  (puis vérifier) :
go version

# 2. Initialiser le projet
cd C:\Users\ktoma\Documents\Projets\P_Go
git init
go mod init github.com/<ton-pseudo>/pgo

# 3. Dépendances
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/huh@latest
go get github.com/charmbracelet/lipgloss@latest
go get golang.org/x/crypto@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/creack/pty@latest
go get golang.org/x/term@latest

# 4. Base de données (dev + tests)
docker run -d --name pgo-postgres -e POSTGRES_PASSWORD=pgo -e POSTGRES_USER=pgo -e POSTGRES_DB=pgo -p 5432:5432 postgres:16
docker exec -it pgo-postgres createdb -U pgo pgo_test
```

> ⚠️ Go n'est **pas** installé sur ta machine : c'est la toute première action.

---

## 11. Check-list « zéro pénalité »

- [ ] Aucun framework web / ORM : uniquement `net/http` + `database/sql`.
- [ ] Aucun composant UI hors Charmbracelet.
- [ ] Le code est **expliquable ligne à ligne** : commentaires sur les choix non triviaux, README qui raconte le pourquoi.
- [ ] Toutes les fonctionnalités de l'énoncé sont présentes (F1→F12).
- [ ] Livrable : **lien Git** accessible + dépôt sur MyGES.
- [ ] Participation de tous les membres à.equal à la soutenance.
- [ ] Démonstration SSH réelle en direct.
- [ ] 15 minutes de questions anticipées.
