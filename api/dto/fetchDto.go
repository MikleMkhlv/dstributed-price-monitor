package dto

type FetchRequest struct {
	Type string `json:"type"`
	// OperationId string   `json:"operationId"`
	Address string   `json:"address"`
	Method  string   `json:"method"`
	Data    []string `json:"data"`
}

type FetchResponce struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (resp *FetchResponce) SetMessage(message error) {
	resp.Message = message.Error()
}
