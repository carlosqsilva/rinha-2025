package handler

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/carlosqsilva/rinha-2025/internal/models"
)

func parseDateQuery(query map[string]string, key string) sql.NullInt64 {
	value := sql.NullInt64{Valid: false}

	if arg, ok := query[key]; ok && arg != "" {
		argTime, err := time.Parse(models.DateLayout, arg[:19])
		if err != nil {
			slog.Debug("failed to parse date", "error", err)
			return value
		}

		if err := value.Scan(argTime.Unix()); err != nil {
			slog.Debug("failed to scan value", "value", argTime.Unix(), "error", err)
			return value
		}

		return value
	}

	return value
}
