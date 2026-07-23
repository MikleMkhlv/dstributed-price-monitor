package fetcher

import (
	"bytes"
	"context"
	"dstributed-price-monitor/internal/fetcher/mapper"
	"dstributed-price-monitor/internal/source"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Client struct {
	ch             chan source.ServiceData
	ErrCh          chan error
	client         http.Client
	monitorAddress string
	mapper         mapper.FetchMaper
}

func NewClient(ch chan source.ServiceData, monitorAddress string) *Client {
	return &Client{
		ch:             ch,
		client:         *http.DefaultClient,
		monitorAddress: monitorAddress,
		mapper:         mapper.FetchMaper{},
	}
}

func (c *Client) SendToMonitor(ctx context.Context) {
	for {
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

			req, err := http.NewRequestWithContext(ctx, "POST", c.monitorAddress, bytes.NewReader(bodyBytes))
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
