CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    identity_provider TEXT NOT NULL,
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    avatar_url TEXT NOT NULL,
    about_me VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE -- soft delete
);

CREATE TABLE IF NOT EXISTS follows (
    follower_id BIGINT NOT NULL,
    following_id BIGINT NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (follower_id, following_id) -- follower_id != following_id
);

CREATE TABLE IF NOT EXISTS subscriptions (
    subscriber_id BIGINT NOT NULL,
    publisher_id BIGINT NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (subscriber_id, publisher_id)
);