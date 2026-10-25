-- Migration 001 : schéma initial de la plateforme P_Go.
-- Convention : un fichier = une version, joué dans l'ordre lexicographique.
-- Les identifiants métier (BSK-…, CMD-…, PDT-…) sont générés par
-- internal/bizid et garantis uniques par une contrainte UNIQUE.

CREATE TABLE users (
    id                BIGSERIAL PRIMARY KEY,
    email             TEXT UNIQUE NOT NULL,
    password_hash     TEXT NOT NULL,
    role              TEXT NOT NULL DEFAULT 'client' CHECK (role IN ('client', 'admin')),
    is_confirmed      BOOLEAN NOT NULL DEFAULT FALSE,
    confirmation_code TEXT UNIQUE,
    reset_code        TEXT,
    reset_expires_at  TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id          BIGSERIAL PRIMARY KEY,
    business_id TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT 'général',
    price_htc   NUMERIC(10, 2) NOT NULL CHECK (price_htc >= 0),
    stock       INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_name     ON products (lower(name));
CREATE INDEX idx_products_category ON products (lower(category));
CREATE INDEX idx_products_price    ON products (price_htc);

CREATE TABLE carts (
    id          BIGSERIAL PRIMARY KEY,
    business_id TEXT UNIQUE NOT NULL,
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'converted', 'abandoned')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Un seul panier actif par utilisateur.
CREATE UNIQUE INDEX idx_carts_active_user ON carts (user_id) WHERE status = 'active';

CREATE TABLE cart_items (
    id             BIGSERIAL PRIMARY KEY,
    cart_id        BIGINT NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    product_id     BIGINT NOT NULL REFERENCES products (id),
    quantity       INT NOT NULL CHECK (quantity > 0),
    unit_price_htc NUMERIC(10, 2) NOT NULL,
    UNIQUE (cart_id, product_id)
);

CREATE TABLE orders (
    id                  BIGSERIAL PRIMARY KEY,
    business_id         TEXT UNIQUE NOT NULL,
    user_id             BIGINT NOT NULL REFERENCES users (id),
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'paid', 'shipped', 'delivered', 'cancelled')),
    cancellation_reason TEXT,
    total_htc           NUMERIC(10, 2) NOT NULL DEFAULT 0,
    total_ttc           NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user ON orders (user_id, created_at DESC);

CREATE TABLE order_items (
    id             BIGSERIAL PRIMARY KEY,
    order_id       BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id     BIGINT REFERENCES products (id) ON DELETE SET NULL,
    business_id    TEXT NOT NULL,
    product_name   TEXT NOT NULL,
    quantity       INT NOT NULL CHECK (quantity > 0),
    unit_price_htc NUMERIC(10, 2) NOT NULL
);

CREATE TABLE payments (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    card_brand TEXT,
    card_last4 TEXT NOT NULL,
    expiry     TEXT NOT NULL,
    amount_ttc NUMERIC(10, 2) NOT NULL,
    status     TEXT NOT NULL DEFAULT 'captured' CHECK (status IN ('captured', 'refused')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order ON payments (order_id);
