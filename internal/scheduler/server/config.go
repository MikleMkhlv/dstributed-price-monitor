package server

import "os"

type MonitorConfig struct {
	AddresServer string
	PortServer   string
}

func NewFetchConfig() *MonitorConfig {
	address := os.Getenv("APPLICATION_ADDRESS")
	port := os.Getenv("APPLICATION_PORT")
	if address == "" || port == "" {
		panic("server.MonitorConfig.NewFetchConfig: APPLICSTION_ADDRESS and APPLICATION_PORT is required")
	}
	return &MonitorConfig{
		AddresServer: address,
		PortServer:   port,
	}
}
