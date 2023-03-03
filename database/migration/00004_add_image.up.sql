CREATE TABLE IF NOT EXISTS attachments
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    image_path TEXT NOT NULL,
    bucket_name TEXT NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS product_attachment
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    attachment_id UUID REFERENCES attachments (id),
    product_id      UUID REFERENCES products (id) NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);