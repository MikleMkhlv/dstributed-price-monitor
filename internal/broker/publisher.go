package broker

import (
	"encoding/json"
	"fmt"

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
	if err != nil {
		return err
	}
	return p.Conn.Publish(subject, payload)
}
