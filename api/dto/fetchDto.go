package dto

type FetchRequest struct {
	Type        string   `json:"type"`
	FullAddress string   `json:"address"`
	Data        []string `json:"data"`
}

type FetchResponce struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (resp *FetchResponce) SetMessage(message error) {
	resp.Message = message.Error()
}
