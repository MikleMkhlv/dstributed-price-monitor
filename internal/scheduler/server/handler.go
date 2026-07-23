package server

import (
	"bytes"
	"context"
	"dstributed-price-monitor/api/dto"
	"dstributed-price-monitor/internal/source"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	MonitorCh chan source.ServiceData
}

func NewHandler(ch chan source.ServiceData) *Handler {
	return &Handler{
		MonitorCh: ch,
	}
}

func (h *Handler) PrepareMonitorMid() gin.HandlerFunc {
	return func(c *gin.Context) {
		errResp := dto.FetchResponce{Status: "error"}
		// currentOperationID := c.Request.Response.Header.Get("operationId")
		currentOperationID := c.GetHeader("operationId")
		if currentOperationID == "" {
			errResp.SetMessage(fmt.Errorf("server.Handler.PrepareMonitorMid: Header operationId is empty. operationId is required field"))
			c.JSON(http.StatusServiceUnavailable, errResp)
			c.Abort()
			return
		}

		newCtxWithTrace := context.WithValue(c.Request.Context(), "operationId", currentOperationID)
		c.Request.WithContext(newCtxWithTrace)
		c.Next()
	}
}

func (h *Handler) Monitor(c *gin.Context) {
	// TODO
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	log.Printf("Monitor: raw body = %s", string(bodyBytes))

	var req dto.FetchResponce
	errResp := dto.FetchResponce{Status: "error"}
	if err := c.ShouldBindJSON(&req); err != nil {
		// TODO
		log.Printf("Monitor: bind error = %v", err)
		errResp.SetMessage(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}
	soucre, err := mapRespFetchToServiceData(req)
	if err != nil {
		// TODO
		log.Printf("Monitor: map error = %v, req.Message = %q", err, req.Message)
		errResp.SetMessage(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}
	select {
	case <-c.Request.Context().Done():
		return
	case h.MonitorCh <- soucre:
		c.JSON(http.StatusAccepted, dto.FetchResponce{Status: "success", Message: "Processing started"})
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
