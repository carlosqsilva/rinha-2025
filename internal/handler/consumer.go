package handler

import (
	"context"
	"log/slog"

	"github.com/carlosqsilva/rinha-2025/internal/queue"
	"github.com/nats-io/nats.go/jetstream"
)

func (h *Handler) StartConsumers(ctx context.Context) {
	h.initPendingConsumer(ctx)
	if h.cfg.EnableSaveConsumer {
		h.initCompletedConsumer(ctx)
	}
}

func (h *Handler) initCompletedConsumer(ctx context.Context) {
	go func() {
		consumer, err := h.queue.Stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
			Durable:       "CompletedConsumer",
			FilterSubject: queue.CompletedSubject,
			AckPolicy:     jetstream.AckExplicitPolicy,
			// MaxDeliver:    1,
		})

		if err != nil {
			slog.Error("failed to create consumer", "error", err)
			return
		}

		for {
			batch, err := consumer.FetchNoWait(10)
			if err != nil {
				continue
			}

			for msg := range batch.Messages() {
				if err := h.SavePayment(msg.Data()); err != nil {
					slog.Info("failed to save payment", "error", err)
					msg.Nak()
					continue
				}

				msg.Ack()
			}
		}
	}()
}

func (h *Handler) initPendingConsumer(ctx context.Context) {
	for i := range ConcurrencyWorker {
		go func(index int) {
			consumer, err := h.queue.Stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
				Durable:       "PendingConsumer",
				FilterSubject: queue.PendingSubject,
				AckPolicy:     jetstream.AckExplicitPolicy,
			})

			if err != nil {
				slog.Error("failed to create consumer", "error", err)
				return
			}

			for {
				batch, err := consumer.Fetch(1)
				if err != nil {
					continue
				}

				for msg := range batch.Messages() {
					slog.Debug("processing payment", "Consumer", index)
					if err := h.ProcessPayment(msg.Data()); err != nil {
						slog.Error("failed to process payment", "error", err)
						msg.Nak()
						continue
					}

					msg.Ack()
				}
			}
		}(i)
	}
}
