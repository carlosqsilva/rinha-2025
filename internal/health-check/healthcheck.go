package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/carlosqsilva/rinha-2025/internal/config"
	"github.com/carlosqsilva/rinha-2025/internal/handler"
	"github.com/carlosqsilva/rinha-2025/internal/models"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/timeout"
	"github.com/gofiber/fiber/v2"
)

type HealthChecker struct {
	cfg     *config.Config
	handler *handler.Handler
}

func New(cfg *config.Config, handler *handler.Handler) *HealthChecker {
	return &HealthChecker{
		cfg:     cfg,
		handler: handler,
	}
}

func (h *HealthChecker) StartHealthMonitor(ctx context.Context) {
	if h.cfg.EnableHealthCheck {
		go h.initHealthMonitor(ctx)
	}
}

func (h *HealthChecker) initHealthMonitor(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go h.healthMonitor()
		case <-ctx.Done():
			slog.Info("shutting down health monitor")
			return
		}
	}
}

func (h *HealthChecker) healthMonitor() {
	body, err := failsafe.Get(
		func() ([]byte, error) {
			return checkHealth(h.cfg.DefaultUrl)
		},
		timeout.With[[]byte](1*time.Second),
	)

	if err != nil {
		slog.Error("Health check failed", "error", err)
		h.handler.OpenCB()
		return
	}

	var healthStatus models.HealthStatus
	if err := sonic.Unmarshal(body, &healthStatus); err != nil {
		slog.Error("Health check failed to parse response", "error", err)
		return
	}
	slog.Debug("HealthStatus", "isFailing", healthStatus.Failing)

	if healthStatus.Failing {
		h.handler.OpenCB()
	} else {
		h.handler.CloseCB()
	}
}

func checkHealth(url string) ([]byte, error) {
	agent := fiber.Get(fmt.Sprintf("%s/payments/service-health", url))
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != fiber.StatusOK {
		return nil, fmt.Errorf("health check failed with status code: %d", statusCode)
	}

	return body, nil
}
