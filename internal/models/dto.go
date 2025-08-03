package models

import (
	"fmt"
	"reflect"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/bytedance/sonic"
)

type CreatePaymentRequest struct {
	CorrelationId string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

type CreatePayment struct {
	CorrelationId string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
	RequestedAt   string  `json:"requestedAt"`
}

type PaymentSummaryInfo struct {
	TotalRequest uint64  `json:"totalRequests"`
	TotalAmount  float64 `json:"totalAmount"`
}

type PaymentSummary struct {
	Default  PaymentSummaryInfo `json:"default"`
	Fallback PaymentSummaryInfo `json:"fallback"`
}

type HealthStatus struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

const (
	DateLayout = "2006-01-02T15:04:05"
)

func init() {
	sonic.Pretouch(reflect.TypeOf(CreatePaymentRequest{}))
	sonic.Pretouch(reflect.TypeOf(CreatePayment{}))
	sonic.Pretouch(reflect.TypeOf(PaymentSummaryInfo{}))
	sonic.Pretouch(reflect.TypeOf(PaymentSummary{}))
	sonic.Pretouch(reflect.TypeOf(HealthStatus{}))
}

func (d *CreatePaymentRequest) Marshal() []byte {
	msg, seg := capnp.NewSingleSegmentMessage(nil)
	book, err := NewRootPaymentData(seg)
	if err != nil {
		panic(fmt.Sprintf("NewPaymentDto: %v", err))
	}

	book.SetCorrelationId(d.CorrelationId)
	book.SetAmount(d.Amount)
	book.SetRequestedAt(time.Now().Format(DateLayout))

	payload, _ := msg.Marshal()
	return payload
}

func CreatePaymentPayload(data []byte) []byte {
	msg, err := capnp.Unmarshal(data)
	if err != nil {
		panic(fmt.Sprintf("capnp.Unmarshal: %v", err))
	}

	payment, err := ReadRootPaymentData(msg)
	if err != nil {
		panic(fmt.Sprintf("ReadRootPaymentDto: %v", err))
	}

	correlationId, err := payment.CorrelationId()
	requestedAt, err := payment.RequestedAt()

	newPayload := &CreatePayment{
		CorrelationId: correlationId,
		Amount:        payment.Amount(),
		RequestedAt:   requestedAt,
	}
	payload, _ := sonic.Marshal(newPayload)
	return payload
}

func UpdatePaymentProcessor(data []byte, processor uint8) []byte {
	msg, err := capnp.Unmarshal(data)
	if err != nil {
		panic(fmt.Sprintf("capnp.Unmarshal: %v", err))
	}

	payment, err := ReadRootPaymentData(msg)
	if err != nil {
		panic(fmt.Sprintf("ReadRootPaymentDto: %v", err))
	}

	payment.SetProcessor(uint8(processor))
	payload, err := msg.Marshal()
	return payload
}

func DecodePaymentData(data []byte) PaymentData {
	msg, _ := capnp.Unmarshal(data)
	payment, _ := ReadRootPaymentData(msg)
	return payment
}
