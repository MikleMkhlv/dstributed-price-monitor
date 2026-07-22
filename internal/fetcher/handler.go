package fetcher

import (
	"context"
	dto "dstributed-price-monitor/api/dto"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	FetchCh chan dto.FetchRequest
}

func NewHandler(ch chan dto.FetchRequest) *Handler {
	return &Handler{
		FetchCh: ch,
	}
}

func (h *Handler) PrepareFetchMid() gin.HandlerFunc {
	return func(c *gin.Context) {
		errResp := dto.FetchResponce{Status: "error"}
		if len(h.FetchCh) == cap(h.FetchCh) {
			errResp.SetMessage(fmt.Errorf("fetcher.Handler.PrepareFetchMid: fetch chanal is crowded"))
			c.JSON(http.StatusServiceUnavailable, errResp)
			c.Abort()
			return
		}
		currentOperationID := c.Request.Response.Header.Get("operationId")
		if currentOperationID == "" {
			errResp.SetMessage(fmt.Errorf("fetcher.Handler.PrepareFetchMid: Header operationId is empty. operationId is required field"))
			c.JSON(http.StatusServiceUnavailable, errResp)
			c.Abort()
			return
		}

		newCtxWithTrace := context.WithValue(c.Request.Context(), "operationId", currentOperationID)
		c.Request.WithContext(newCtxWithTrace)
		c.Next()
	}
}

func (h *Handler) Fetch(c *gin.Context) {
	var req dto.FetchRequest
	errResp := dto.FetchResponce{Status: "error"}
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp.SetMessage(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}
	select {
	case <-c.Request.Context().Done():
		return
	case h.FetchCh <- req:
		c.JSON(http.StatusAccepted, dto.FetchResponce{Status: "success", Message: "Processing started"})
	}
}
