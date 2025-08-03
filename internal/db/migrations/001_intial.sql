-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payments (
    correlation_id TEXT PRIMARY KEY,
    amount INTEGER NOT NULL,
    requested_at INTEGER NOT NULL,
    processor INTEGER NOT NULL
);

CREATE INDEX idx_payments_requested_at ON payments (requested_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payments_requested_at;
DROP TABLE IF EXISTS payments;
-- +goose StatementEnd
