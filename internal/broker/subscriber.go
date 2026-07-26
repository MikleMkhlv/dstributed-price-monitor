package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

type Subscriber[T any] struct {
	conn *nats.Conn
	data chan T
	sub  *nats.Subscription
	wg   sync.WaitGroup
}

func NewSubscription[T any](conn *Nats, inputCh chan T) (*Subscriber[T], error) {
	if conn.Conn == nil {
		return nil, fmt.Errorf("Subscription error. No connection to the broker.")
	}
	return &Subscriber[T]{
		data: inputCh,
		conn: conn.Conn,
		wg:   sync.WaitGroup{},
	}, nil
}

func (r *Subscriber[T]) Start(subject string) error {
	sub, err := r.conn.Subscribe(subject, func(msg *nats.Msg) {
		r.wg.Add(1)
		defer r.wg.Done()
		var event T
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("broker.Subscriber.Start: error unmarshal json. %v", err)
			return
		}
		r.data <- event
	})
	if err != nil {
		return err
	}
	r.sub = sub
	return nil
}

func (r *Subscriber[T]) Stop() error {
	if err := r.sub.Unsubscribe(); err != nil {
		return err
	}
	r.wg.Wait()
	close(r.data)
	return nil
}
