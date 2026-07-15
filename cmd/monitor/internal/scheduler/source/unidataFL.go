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

type UnidataFLSource struct {
	url    string
	method string
	mdmIds []string
	client *http.Client
	wg     sync.WaitGroup
}

// func (c *UnidataFLSource) isSourceRecord() {}

func NewUnidataFLSource(url string, method string, data []string) (*UnidataFLSource, error) {
	if url == "" {
		return nil, fmt.Errorf("source.NewUnidataFLSource: url must not be empty")
	}
	if method == "" {
		return nil, fmt.Errorf("source.NewUnidataFLSource: method must not be empty")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("source.NewUnidataFLSource: data must not be empty")
	}
	return &UnidataFLSource{
		url:    url,
		method: method,
		mdmIds: data,
		client: &http.Client{},
		wg:     sync.WaitGroup{},
	}, nil
}

func (c *UnidataFLSource) Pool(ctx context.Context, outCh chan ServiceData, errCh chan error) {
	for _, id := range c.mdmIds {
		func() {
			select {
			case <-ctx.Done():
				errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel")
				return
			default:
			}

			errSend := func(ctx context.Context, errMessage error) {
				select {
				case <-ctx.Done():
					errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel")
					return
				case errCh <- errMessage:
				}
			}
			req, err := http.NewRequestWithContext(ctx, "GET", buildFullUrl(c.url, c.method, id), nil)
			if err != nil {
				errSend(ctx, fmt.Errorf("source.UnidataFLSource.Pool: error pool request: %v", err))
				return
			}
			log.Printf("source.UnidataFLSource.Pool: send request in %s", buildFullUrl(c.url, c.method, id))
			resp, err := c.client.Do(req)
			if err != nil {
				errSend(ctx, fmt.Errorf("source.UnidataFLSource.Pool: error send request: %v", err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errSend(ctx, fmt.Errorf("source.UnidataFLSource.Poll: unexpected status code: %d", resp.StatusCode))
				return
			}
			log.Printf("source.Pool: status code - %d", resp.StatusCode)

			var citizens []Citizen
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&citizens); err != nil {
				errSend(ctx, fmt.Errorf("source.UnidataFLSource.Pool: error decode request body: %v", err))
				return
			}
			if len(citizens) == 0 {
				log.Printf("source.UnidataFLSource.Pool: citizen %s does not exist", id)
				select {
				case <-ctx.Done():
					errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel")
				case errCh <- fmt.Errorf("source.UnidataFLSource.Pool: citizen %s does not exist", id):
					return
				}

			}
			// log.Printf("source.Pool: decode citizen complete. %s", citizens[0].String())

			select {
			case <-ctx.Done():
				errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel")
				return
			case outCh <- citizens[0]:
			}
		}()
	}
}

func buildFullUrl(address, method, id string) string {
	var fullUrl strings.Builder
	fullUrl.WriteString(address)
	fullUrl.WriteString(method)
	fullUrl.WriteString(id)
	return fullUrl.String()
}

type Citizen struct {
	Data CitizenData `json:"citizen"`
}

type CitizenData struct {
	LastName       string `json:"last_name"`
	FirstName      string `json:"given_name_one"`
	MiddleName     string `json:"given_name_two"`
	ValidationMask int    `json:"validationmask"`
}

func (ctz Citizen) IsImpl() {}

func (ctz Citizen) String() string {
	var citizen strings.Builder
	fmt.Fprintf(&citizen, "firstName:%s, lastName:%s, middleName:%s, valMask:%d\n", ctz.Data.FirstName,
		ctz.Data.LastName, ctz.Data.MiddleName, ctz.Data.ValidationMask)
	return citizen.String()
}
