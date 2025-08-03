package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/carlosqsilva/rinha-2025/internal/api"
	"github.com/carlosqsilva/rinha-2025/internal/config"
	"github.com/carlosqsilva/rinha-2025/internal/db"
	"github.com/carlosqsilva/rinha-2025/internal/handler"
	healthcheck "github.com/carlosqsilva/rinha-2025/internal/health-check"
	"github.com/carlosqsilva/rinha-2025/internal/queue"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(config.New),
		fx.Provide(db.Connect),
		fx.Provide(queue.New),
		fx.Provide(handler.New),
		fx.Provide(healthcheck.New),
		fx.Provide(api.New),
		fx.Invoke(func(
			lc fx.Lifecycle,
			api *api.Api,
			queue *queue.Message,
			handler *handler.Handler,
			health *healthcheck.HealthChecker,
			cfg *config.Config,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					health.StartHealthMonitor(ctx)
					handler.StartConsumers(ctx)
					api.Listen()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					api.Close(ctx)
					queue.Close(ctx)
					return nil
				},
			})
		}),
	)

	ctx := context.Background()

	if err := app.Start(ctx); err != nil {
		slog.Error("failed to start application", "error", err)
	}

	<-app.Done()
	slog.Info("Shutdown signal received, stopping application...")

	stopCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		slog.Error("failed to stop application gracefully", "error", err)
	} else {
		slog.Info("Application stopped gracefully")
	}
}
