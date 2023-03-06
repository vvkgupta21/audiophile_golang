CREATE TYPE order_status AS ENUM (
    'ordered',
    'shipping',
    'delivered'
    );

CREATE TABLE IF NOT EXISTS orders
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    cart_id      UUID REFERENCES carts (id) NOT NULL,
    order_status order_status,
    address_id   UUID REFERENCES user_addresses (id),
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
