package queue

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/carlosqsilva/rinha-2025/internal/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	StreamName              = "PAYMENTS"
	PendingSubject   string = "payment.pending"
	CompletedSubject string = "payment.completed"
)

type Message struct {
	Coon *nats.Conn
	// KV            jetstream.KeyValue
	Stream        jetstream.Stream
	StreamManager jetstream.JetStream
}

func New(cfg *config.Config) *Message {
	nc, err := nats.Connect(cfg.NatsUrl)
	if err != nil {
		slog.Error("failed to connect to nats server", "error", err)
		os.Exit(1)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		slog.Error("failed to create jetstream instance", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
	// 	Bucket: "profiles",
	// })
	if err != nil {
		slog.Error("failed to create key value store", "error", err)
		os.Exit(1)
	}

	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      StreamName,
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{"payment.>"},
	})
	if err != nil {
		slog.Error("failed to create stream", "error", err)
		os.Exit(1)
	}

	return &Message{
		Coon: nc,
		// KV:            kv,
		Stream:        stream,
		StreamManager: js,
	}
}

func (m *Message) AddPending(payload []byte) {
	if _, err := m.StreamManager.PublishAsync(PendingSubject, payload); err != nil {
		slog.Error("failed to publish message", "Subject", PendingSubject, "error", err)
	}
}

func (m *Message) AddCompleted(payload []byte) {
	if _, err := m.StreamManager.PublishAsync(CompletedSubject, payload); err != nil {
		slog.Error("failed to publish message", "Subject", CompletedSubject, "error", err)
	}
}

func (m *Message) Close(ctx context.Context) {
	slog.Info("cleaning up queue stream")
	m.Stream.Purge(ctx, jetstream.WithPurgeKeep(0))
	m.Coon.Drain()
	m.Coon.Close()
}
