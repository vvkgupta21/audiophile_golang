CREATE TABLE IF NOT EXISTS users
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    email       TEXT NOT NULL,
    password    TEXT NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    update_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TYPE category AS ENUM (
    'headphones',
    'speakers',
    'earphones'
    );

CREATE TABLE IF NOT EXISTS products
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name         TEXT     NOT NULL,
    price        DECIMAL  NOT NULL,
    description  TEXT,
    is_available BOOLEAN  NOT NULL,
    quantity     INTEGER  NOT NULL,
    category     category NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    update_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at  TIMESTAMP WITH TIME ZONE
);

CREATE TYPE cart_status AS ENUM (
    'active',
    'inactive'
    );

CREATE TABLE IF NOT EXISTS carts
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users (id) NOT NULL,
    status     cart_status                NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    update_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    cart_id    UUID REFERENCES carts (id) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE address_type AS ENUM (
    'home',
    'office',
    'other'
    );

CREATE TABLE IF NOT EXISTS user_addresses
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users (id) NOT NULL,
    address      TEXT                       NOT NULL,
    address_type address_type               NOT NULL,
    lat          DECIMAL                    NOT NULL,
    long         DECIMAL                    NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT Now(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT Now(),
    archived_at  TIMESTAMP WITH TIME ZONE
);

CREATE TYPE role_type AS ENUM (
    'admin',
    'user'
    );

CREATE TABLE IF NOT EXISTS user_roles
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users (id) NOT NULL,
    role        role_type                  NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT Now(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS cart_products
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    cart_id    UUID REFERENCES carts (id)    NOT NULL,
    product_id UUID REFERENCES products (id) NOT NULL,
    quantity   INTEGER                       NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT Now()
);