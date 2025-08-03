-- name: CreatePayment :exec
INSERT INTO payments (
    correlation_id, amount, requested_at, processor
) VALUES (
    ?, ?, ?, ?
);
