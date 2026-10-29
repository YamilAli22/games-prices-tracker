CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    steam_app_id INT NOT NULL,
    price NUMERIC NOT NULL,
    is_free BOOLEAN NOT NULL DEFAULT false,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_price_history_app_id ON price_history(steam_app_id, checked_at DESC);
