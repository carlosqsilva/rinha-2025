package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func handlePayment(url string, payload []byte) error {
	agent := fiber.Post(fmt.Sprintf("%s/payments", url))
	agent.Request().Header.SetContentType(fiber.MIMEApplicationJSON)
	agent.Body(payload)

	statusCode, _, errs := agent.Bytes()
	if len(errs) > 0 {
		return fmt.Errorf("%s: %w", errs[0], ErrCreatePayment)
	}

	if statusCode != fiber.StatusOK {
		return fmt.Errorf("Invalid response status %d: %w", statusCode, ErrCreatePayment)
	}

	return nil
}
