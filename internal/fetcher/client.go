package fetcher

import (
	"bytes"
	"context"
	"dstributed-price-monitor/config"
	"dstributed-price-monitor/internal/fetcher/mapper"
	"dstributed-price-monitor/internal/source"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Client struct {
	ch             chan source.ServiceData
	ErrCh          chan error
	client         http.Client
	monitorAddress string
	mapper         mapper.FetchMaper
	cfg            *config.Config
}

func NewClient(ch chan source.ServiceData, cfg *config.Config) *Client {
	return &Client{
		ch: ch,
		// TODO: канал err надо передавать в аргументах. Пока как загдушка будет создаваться буфферизированный канал
		ErrCh:          make(chan error, 100),
		client:         *http.DefaultClient,
		monitorAddress: cfg.General.MonitorAddress,
		mapper:         mapper.FetchMaper{},
		cfg:            cfg,
	}
}

func (c *Client) SendToMonitor(ctx context.Context) {
	for {
		operationID := uuid.New().String()
		select {
		case <-ctx.Done():
			log.Print("fetcher.Client.SendToMonitor. context cancel")
			return
		case res, ok := <-c.ch:
			if !ok {
				select {
				case <-ctx.Done():
					log.Print("fetcher.Client.SendToMonitor. context cancel")
					return
				case c.ErrCh <- fmt.Errorf("fetcher.Client.SendToMonitor: channel closed"):
					return
				}
			}
			data, err := c.mapper.CitizenToFetchResponse(res)
			bodyBytes, err := json.Marshal(data)
			if err != nil {
				select {
				case <-ctx.Done():
					log.Print("fetcher.Client.SendToMonitor. context cancel")
					return
				case c.ErrCh <- err:
					continue
				}
			}
			fullAdd := fmt.Sprintf("%s/api/monitor", c.monitorAddress)
			req, err := http.NewRequestWithContext(ctx, "POST", fullAdd, bytes.NewReader(bodyBytes))
			if err != nil {
				errMsg := fmt.Errorf("fetcher.Client.SendToMonitor: error create request. %v", err)
				select {
				case <-ctx.Done():
					log.Print("fetcher.Client.SendToMonitor. context cancel")
					return
				case c.ErrCh <- errMsg:
					continue
				}
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("operationId", operationID)

			resp, err := c.client.Do(req)
			if err != nil {
				errMsg := fmt.Errorf("fetcher.Client.SendToMonitor: error send request. %v", err)
				select {
				case <-ctx.Done():
					log.Print("fetcher.Client.SendToMonitor. context cancel")
					return
				case c.ErrCh <- errMsg:
					continue
				}
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				errMsg := fmt.Errorf("fetcher.Client.SendToMonitor: bad status: %s", resp.Status)
				select {
				case <-ctx.Done():
					log.Print("fetcher.Client.SendToMonitor. context cancel")
					return
				case c.ErrCh <- errMsg:
					continue
				}
			}
			log.Print("fetcher.Client.SendToMonitor: send request in monitor is successful")
		}
	}
}
