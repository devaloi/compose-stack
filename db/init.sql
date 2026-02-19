CREATE TABLE IF NOT EXISTS items (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO items (name, description) VALUES
    ('Example Item', 'This is a seed item created on first startup'),
    ('Another Item', 'A second seed item for demonstration')
ON CONFLICT DO NOTHING;
