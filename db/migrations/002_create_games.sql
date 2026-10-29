CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    steam_app_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    current_price NUMERIC(10, 2),
    target_price NUMERIC(10, 2) NOT NULL DEFAULT 0,
    is_free BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT games_user_game_unique UNIQUE (user_id, steam_app_id)
);
