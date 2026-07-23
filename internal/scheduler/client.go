package scheduler

import (
	"bytes"
	"context"
	"dstributed-price-monitor/config"
	"dstributed-price-monitor/internal/scheduler/mapper"
	"dstributed-price-monitor/internal/source"
	"dstributed-price-monitor/internal/worker/semaphore"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Client struct {
	InputCh        chan source.Record
	ErrCh          chan error
	client         http.Client
	semaphore      semaphore.Semaphore
	fetcherAddress string
	mapper         mapper.MapperShd
	ctx            *config.Config
}

func NewClient(inputCh chan source.Record, errCh chan error, ctx *config.Config) *Client {
	return &Client{
		InputCh:        inputCh,
		ErrCh:          errCh,
		client:         *http.DefaultClient,
		fetcherAddress: ctx.General.FetchAddress,
		mapper:         mapper.MapperShd{},
		ctx:            ctx,
	}
}

func (c *Client) SendToFetch(ctx context.Context) {
	for {
		operationID := uuid.New().String()
		select {
		case <-ctx.Done():
			log.Print("scheduler.Client.SendToFetch: context cancel")
			return
		case res, ok := <-c.InputCh:
			if !ok {
				select {
				case <-ctx.Done():
					log.Print("scheduler.Client.SendToFetch: context cancel")
					return
				case c.ErrCh <- fmt.Errorf("scheduler.Client.SendToFetch: channel closed"):
					return
				}
			}
			data := c.mapper.RecordToRequestForFether(res)
			bodyBytes, err := json.Marshal(data)
			if err != nil {
				select {
				case <-ctx.Done():
					log.Print("scheduler.Client.SendToFetch: context cancel")
					return
				case c.ErrCh <- err:
					continue
				}
			}
			fullAdd := fmt.Sprintf("%s/api/fetch", c.fetcherAddress)
			req, err := http.NewRequestWithContext(ctx, "POST", fullAdd, bytes.NewReader(bodyBytes))
			if err != nil {
				errMsg := fmt.Errorf("scheduler.Client.SendToFetch: error create request. %v", err)
				select {
				case <-ctx.Done():
					log.Print("scheduler.Client.SendToFetch: context cancel")
					return
				case c.ErrCh <- errMsg:
					continue
				}
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("operationId", operationID)

			resp, err := c.client.Do(req)
			if err != nil {
				errMsg := fmt.Errorf("scheduler.Client.SendToFetch: error send request. %v", err)
				select {
				case <-ctx.Done():
					log.Print("scheduler.Client.SendToFetch: context cancel")
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
					log.Print("scheduler.Client.SendToFetch: context cancel")
					return
				case c.ErrCh <- errMsg:
					continue
				}
			}
			log.Print("scheduler.Client.SendToFetch: send request in monitor is successful")
		}
	}
}
