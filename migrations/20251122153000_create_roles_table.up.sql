BEGIN;

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_roles_name ON roles(name);

INSERT INTO roles (id, name, description) VALUES 
    (1, 'user', 'Standard user with basic permissions'),
    (2, 'admin', 'Administrator with full system access')
ON CONFLICT (name) DO NOTHING;

COMMIT;
