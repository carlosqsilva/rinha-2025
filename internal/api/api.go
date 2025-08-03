package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/bytedance/sonic"
	"github.com/carlosqsilva/rinha-2025/internal/config"
	"github.com/carlosqsilva/rinha-2025/internal/handler"
	"github.com/gofiber/fiber/v2"
)

type Api struct {
	app      *fiber.App
	cfg      *config.Config
	listener net.Listener
}

func New(cfg *config.Config, handler *handler.Handler) *Api {
	socketPath := fmt.Sprintf("./sockets/%s", cfg.Socket)
	slog.Info("Creating Unix socket", "path", socketPath)

	// Ensure the directory exists
	// socketDir := filepath.Dir(cfg.Socket)
	// if err := os.MkdirAll(socketDir, 0755); err != nil {
	// 	slog.Error("Failed to create socket directory", "error", err)
	// 	panic(err)
	// }
	// Remove existing socket file if it exists
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		slog.Error("Failed to remove existing socket", "error", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		slog.Error("Failed to listen on Unix socket", "error", err)
	}

	if err := os.Chmod(socketPath, 0766); err != nil {
		slog.Error("Failed to set permissions for", "socket", socketPath, "error", err)
	}

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Post("/payments", handler.CreatePayment)
	app.Get("/payments-summary", handler.GetSummary)

	return &Api{
		app:      app,
		cfg:      cfg,
		listener: listener,
	}
}

func (a *Api) Listen() {
	go func() {
		if err := a.app.Listener(a.listener); err != nil {
			slog.Error("Server failed to start", "error", err)
		}
	}()
}

func (a *Api) Close(ctx context.Context) {
	slog.Error("Server shutting down...")
	if err := a.app.ShutdownWithContext(ctx); err != nil {
		slog.Error("API shutdown failed", "error", err)
	}
	if err := a.listener.Close(); err != nil {
		slog.Error("Listener shutdown failed", "error", err)
	}
	if err := os.Remove(a.cfg.Socket); err != nil {
		slog.Error("Failed to cleanup socket", "error", err)
	}
}
