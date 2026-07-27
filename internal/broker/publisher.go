package broker

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/nats-io/nats.go"
)

type Publisher[T any] struct {
	Conn *nats.Conn
}

func NewPublisher[T any](conn *Nats) (*Publisher[T], error) {
	if conn == nil {
		return nil, fmt.Errorf("publisher creation failed: no connection to broker")
	}
	return &Publisher[T]{
		Conn: conn.Conn,
	}, nil
}

func (p *Publisher[T]) Publish(subject string, data T) error {
	payload, err := json.Marshal(data)
	log.Printf("DEBUG: pub payload = %s", string(payload))
	if err != nil {
		return err
	}
	msg := &nats.Msg{
		Subject: subject,
		Data:    payload,
		Header:  nats.Header{},
	}
	operationID := uuid.New().String()
	msg.Header.Set("operationId", operationID)
	return p.Conn.PublishMsg(msg)
}
