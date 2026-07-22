package fetcher

import "os"

type FetchConfig struct {
	AddresServer string
	PortServer   string
}

func NewFetchConfig() *FetchConfig {
	address := os.Getenv("APPLICATION_ADDRESS")
	port := os.Getenv("APPLICATION_PORT")
	if address == "" || port == "" {
		panic("fetcher.FetchConfig.NewFetchConfig: APPLICSTION_ADDRESS and APPLICATION_PORT is required")
	}
	return &FetchConfig{
		AddresServer: address,
		PortServer:   port,
	}
}
