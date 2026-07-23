package fetcher

import (
	"context"
	"dstributed-price-monitor/internal/source"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Address  string
	Port     string
	Handlers *Handler
	conf     *FetchConfig
	router   *gin.Engine
	httpSrv  *http.Server
}

func NewServer(fetchCh chan source.Record, cfg *FetchConfig) *Server {
	addr := fmt.Sprintf("%s:%s", cfg.AddresServer, cfg.PortServer)
	router := gin.Default()
	return &Server{
		Address:  cfg.AddresServer,
		Port:     cfg.PortServer,
		Handlers: NewHandler(fetchCh),
		conf:     cfg,
		router:   router,
		httpSrv: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Server) RunServer() error {
	s.router.POST("/api/fetch", s.Handlers.PrepareFetchMid(), s.Handlers.Fetch)

	addr := fmt.Sprintf("%s:%s", s.Address, s.Port)
	log.Printf("fetcher.Server.RunServer: server run on %s", addr)

	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	if s.httpSrv == nil {
		return nil
	}
	log.Println("fetcher.Server.Stop: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpSrv.Shutdown(ctx)
}
