-- +goose Up
CREATE TABLE feed_follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL, 
    user_id TEXT NOT NULL UNIQUE,
    CONSTRAINT fk_user_id
    FOREIGN KEY (user_id)
    REFERENCES users(name) ON DELETE CASCADE,
    feed_id UUID NOT NULL UNIQUE,
    CONSTRAINT fk_feed_id
    FOREIGN KEY (feed_id)
    REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feed_follows;