-- name: AggregatePaymentsByProcessorAndDateRange :many
SELECT
  processor,
  COUNT(*) AS total_count,
  SUM(amount) AS total_amount
FROM
  payments
WHERE
  (requested_at >= sqlc.narg('from') OR sqlc.narg('from') IS NULL)
  AND (requested_at <= sqlc.narg('to') OR sqlc.narg('to') IS NULL)
GROUP BY
  processor;
