package broker

import (
	"dstributed-price-monitor/config"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

type Nats struct {
	Conn *nats.Conn
}

func NewNats(cfg *config.Config) (*Nats, error) {
	addr := cfg.Nats.Address
	if addr == "" {
		addr = nats.DefaultURL
	}

	strings.TrimSpace(addr)
	ns, err := nats.Connect(
		addr,
		nats.Timeout(time.Second*time.Duration(cfg.Nats.Timeout)),
	)
	if err != nil {
		return nil, fmt.Errorf("error connect to nuts(%s). %v", addr, err)
	}

	log.Print("broker.Nats.NewNats: connections with nuts is sucessfull:", ns.Status())
	return &Nats{
		Conn: ns,
	}, nil
}

func (n *Nats) Close() error {
	if n.Conn == nil {
		return fmt.Errorf("The nuts connections had already been broken.")
	}
	n.Conn.Close()
	fmt.Print("broker.Nats.NewNats: connections with nuts is close")
	return nil
}
