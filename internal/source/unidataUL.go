package source

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

type UnidataULSource struct {
	url    string
	method string
	mdmIds []string
	client *http.Client
	wg     sync.WaitGroup
}

func (c *UnidataULSource) isSourceRecord() {}

func NewUnidataULSource(url string, method string, data []string) (*UnidataULSource, error) {
	if url == "" {
		return nil, fmt.Errorf("source.UnidataULSource.NewUnidataULSource: url must not be empty")
	}
	if method == "" {
		return nil, fmt.Errorf("source.UnidataULSource.NewUnidataULSource: method must not be empty")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("source.UnidataULSource.NewUnidataULSource: data must not be empty")
	}
	return &UnidataULSource{
		url:    url,
		method: method,
		mdmIds: data,
		client: &http.Client{},
		wg:     sync.WaitGroup{},
	}, nil
}

func (c *UnidataULSource) Pool(ctx context.Context, outCh chan ServiceData, errCh chan error) {
	for _, id := range c.mdmIds {
		func() {
			select {
			case <-ctx.Done():
				errCh <- fmt.Errorf("source.UnidataULSource.Pool: context cancel")
				return
			default:
			}

			errSend := func(ctx context.Context, errMessage error) {
				select {
				case <-ctx.Done():
					errCh <- fmt.Errorf("source.UnidataULSource.Pool: context cancel")
					return
				case errCh <- errMessage:
				}
			}
			req, err := http.NewRequestWithContext(ctx, "GET", buildFullUrl(c.url, c.method, id), nil)
			if err != nil {
				errSend(ctx, fmt.Errorf("source.UnidataULSource.Pool: error pool request: %v", err))
				return
			}
			log.Printf("source.UnidataULSource.Pool: send request in %s", buildFullUrl(c.url, c.method, id))
			resp, err := c.client.Do(req)
			if err != nil {
				errSend(ctx, fmt.Errorf("source.UnidataULSource.Pool: error send request: %v", err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errSend(ctx, fmt.Errorf("source.UnidataULSource.Poll: unexpected status code: %d", resp.StatusCode))
				return
			}
			log.Printf("source.UnidataULSource.Pool: status code - %d", resp.StatusCode)

			var org []Organization
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&org); err != nil {
				errSend(ctx, fmt.Errorf("source.UnidataULSource.Pool: error decode request body: %v", err))
				return
			}
			if len(org) == 0 {
				log.Printf("source.UnidataULSource.Pool: org %s does not exist", id)
				select {
				case <-ctx.Done():
					errCh <- fmt.Errorf("source.UnidataULSource.Pool: context cancel")
				case errCh <- fmt.Errorf("source.UnidataULSource.Pool: org %s does not exist", id):
					return
				}

			}

			select {
			case <-ctx.Done():
				errCh <- fmt.Errorf("source.UnidataULSource.Pool: context cancel")
				return
			case outCh <- org[0]:
			}
		}()
	}
}

type Organization struct {
	Data OrganizationData `json:"org"`
}

type OrganizationData struct {
	HeadFirstName  string `json:"head_given_name_one"`
	HeadLastName   string `json:"head_last_name"`
	HeadMiddleName string `json:"head_given_name_two"`
	Position       string `json:"head_position"`
}

func (ctz Organization) IsImpl() {}

func (ctz Organization) String() string {
	var citizen strings.Builder
	fmt.Fprintf(&citizen, "HeadfirstName:%s, HeadLastName:%s, HeadMiddleName:%s, valMask:%d\n", ctz.Data.HeadFirstName,
		ctz.Data.HeadLastName, ctz.Data.HeadMiddleName, ctz.Data.Position)
	return citizen.String()
}
