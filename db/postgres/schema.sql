CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    identity_provider TEXT NOT NULL,
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    avatar_url TEXT NOT NULL,
    about_me VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL, -- soft delete
    UNIQUE (email, identity_provider)
);