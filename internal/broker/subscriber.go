package broker

import (
	"dstributed-price-monitor/api/dto"
	"dstributed-price-monitor/internal/fetcher/mapper"
	"dstributed-price-monitor/internal/source"
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

func (r *Subscriber[T]) Start(subject string, mode string) error {
	sub, err := r.conn.Subscribe(subject, func(msg *nats.Msg) {
		r.wg.Add(1)
		defer r.wg.Done()
		switch mode {
		case "fromMonitor":
			var req dto.FetchRequest
			if err := json.Unmarshal(msg.Data, &req); err != nil {
				log.Printf("broker.Subscriber.Start: error unmarshal json[data:=%s]. %v", string(msg.Data), err)
				return
			}
			rec, err := mapFetchToRecord(req)
			if err != nil {
				log.Printf("broker.Subscriber.Start: error map req in struct. %v", err)
				return
			}
			concreteVal, ok := rec.(T)
			if !ok {
				log.Printf("broker.Subscriber.Start: unknown type, got %T", rec)
				return
			}
			select {
			case r.data <- concreteVal:
			default:
				log.Printf("broker.Subscriber.Start: r.data channel is full or blocked, dropping message")
			}
		case "fromFetc":
			var req dto.FetchResponce
			if err := json.Unmarshal(msg.Data, &req); err != nil {
				log.Printf("broker.Subscriber.Start: error unmarshal json[data:=%s]. %v", string(msg.Data), err)
				return
			}
			soucre, err := mapRespFetchToServiceData(req)
			log.Printf("broker.Subscriber.Start.DEBUG: sourse after map. %v", soucre)
			if err != nil {
				log.Printf("broker.Subscriber.Start: map error = %v, req.Message = %q", err, req.Message)
			}
			if concreteVal, ok := soucre.(T); ok {
				r.data <- concreteVal
			} else {
				log.Printf("broker.Subscriber.Start: uncnown type in interface Record")
				return
			}

		}
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
	return nil
}

func mapFetchToRecord(data dto.FetchRequest) (source.Record, error) {
	var mapper mapper.FetchMaper
	switch d := data.Type; d {
	case "unidata_fl":
		res, err := mapper.FetchRequestToUnidataFLSource(data)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "unidata_ul":
		res, err := mapper.FetchRequestToUnidataULSource(data)
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		errMsg := fmt.Errorf("fetcher.Handler.mapFetchToRecord: unknown source type. type = %s", d)
		return nil, errMsg
	}
}

func mapRespFetchToServiceData(data dto.FetchResponce) (source.ServiceData, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data.Message), &raw); err != nil {
		return nil, fmt.Errorf("server.Handler.mapRespFetchToServiceData: error parsing json %w", err)
	}
	switch {
	case hasKey(raw, "citizen"):
		var citizen source.Citizen
		if err := json.Unmarshal([]byte(data.Message), &citizen); err != nil {
			return nil, fmt.Errorf("server.Handler.mapRespFetchToServiceData: error mapping Citizen: %w", err)
		}
		return citizen, nil
	case hasKey(raw, "org"):
		var org source.Organization
		if err := json.Unmarshal([]byte(data.Message), &org); err != nil {
			return nil, fmt.Errorf("server.Handler.mapRespFetchToServiceData: error mapping Citizen: %w", err)
		}
		return org, nil
	default:
		return nil, fmt.Errorf("server.Handler.mapRespFetchToServiceData: unknown service type")
	}
}

func hasKey(raw map[string]json.RawMessage, key string) bool {
	_, ok := raw[key]
	return ok
}
