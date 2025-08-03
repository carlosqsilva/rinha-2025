package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PaymentClient struct {
	client *http.Client
}

func NewPaymentClient() *PaymentClient {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     90 * time.Second,

		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		DisableKeepAlives: false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	return &PaymentClient{
		client: client,
	}
}

func (p *PaymentClient) HandlePayment(baseUrl string, payload []byte) error {
	client := p.client

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/payments", baseUrl), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", ErrCreatePayment)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", ErrCreatePayment)
	}

	defer func() {
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	return nil
}
