package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/carlosqsilva/rinha-2025/internal/config"
	"github.com/carlosqsilva/rinha-2025/internal/db"
	"github.com/carlosqsilva/rinha-2025/internal/models"
	"github.com/carlosqsilva/rinha-2025/internal/queue"
	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/fallback"
	"github.com/failsafe-go/failsafe-go/timeout"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	db      *db.Queries
	cfg     *config.Config
	queue   *queue.Message
	breaker circuitbreaker.CircuitBreaker[ProcessorType]
	timeout timeout.Timeout[ProcessorType]
}

const (
	ConcurrencyWorker int = 8
)

var ErrCreatePayment error = errors.New("failed to process payment")

func New(queue *queue.Message, db *db.Queries, cfg *config.Config) *Handler {
	timeoutCfg := timeout.With[ProcessorType](1 * time.Second)
	breakerCfg := circuitbreaker.Builder[ProcessorType]().
		HandleErrors(ErrCreatePayment, timeout.ErrExceeded).
		WithFailureThreshold(2).
		WithDelay(5 * time.Second).
		WithSuccessThreshold(2).
		Build()

	return &Handler{
		db:      db,
		cfg:     cfg,
		queue:   queue,
		breaker: breakerCfg,
		timeout: timeoutCfg,
	}
}

func (h *Handler) CreatePayment(c *fiber.Ctx) error {
	var req models.CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("failed to parse request body", "error", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	h.queue.AddPending(req.Marshal())

	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) GetSummary(c *fiber.Ctx) error {
	queries := c.Queries()
	fromArg := parseDateQuery(queries, "from")
	toArg := parseDateQuery(queries, "to")

	results, err := h.db.AggregatePaymentsByProcessorAndDateRange(c.Context(), db.AggregatePaymentsByProcessorAndDateRangeParams{
		From: fromArg,
		To:   toArg,
	})

	var summary models.PaymentSummary
	if err != nil {
		slog.Error("failed to get summary", "error", err)
		return c.Status(fiber.StatusOK).JSON(summary)
	}

	for _, result := range results {
		switch uint8(result.Processor) {
		case ProcessorDefault:
			summary.Default.TotalAmount = float64(result.TotalAmount.Float64) / 100
			summary.Default.TotalRequest = uint64(result.TotalCount)
		case ProcessorFallback:
			summary.Fallback.TotalAmount = float64(result.TotalAmount.Float64) / 100
			summary.Fallback.TotalRequest = uint64(result.TotalCount)
		default:
			continue
		}
	}

	return c.Status(fiber.StatusOK).JSON(summary)
}

func (h *Handler) ProcessPayment(data []byte) error {
	payload := models.CreatePaymentPayload(data)

	processor, err := failsafe.Get(
		func() (ProcessorType, error) {
			slog.Info("using default")
			err := handlePayment(h.cfg.DefaultUrl, payload)
			return ProcessorDefault, err
		},
		fallback.WithFunc(func(e failsafe.Execution[ProcessorType]) (ProcessorType, error) {
			slog.Info("using fallback")
			err := handlePayment(h.cfg.FallbackUrl, payload)
			return ProcessorFallback, err
		}),
		h.breaker,
		h.timeout,
	)

	if err != nil {
		return err
	}

	if processor == ProcessorFallback {
		data = models.UpdatePaymentProcessor(data, processor)
	}

	h.queue.AddCompleted(data)
	return nil
}

func (h *Handler) SavePayment(data []byte) error {
	dto := models.DecodePaymentData(data)
	correlationId, _ := dto.CorrelationId()
	amount := int64(dto.Amount() * 100)
	requestedAtStr, _ := dto.RequestedAt()
	requestedAt, _ := time.Parse(models.DateLayout, requestedAtStr)
	processor := dto.Processor()

	ctx := context.Background()
	err := h.db.CreatePayment(ctx, db.CreatePaymentParams{
		CorrelationID: correlationId,
		Amount:        amount,
		RequestedAt:   requestedAt.Unix(),
		Processor:     int64(processor),
	})
	if err != nil {
		slog.Error("failed to persist to db", "error", err)
	}

	return err
}

func (h *Handler) CloseCB() {
	h.breaker.Close()
}

func (h *Handler) OpenCB() {
	h.breaker.Open()
}
