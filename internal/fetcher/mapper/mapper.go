package mapper

import (
	dto "dstributed-price-monitor/api/dto"
	"dstributed-price-monitor/internal/source"
	"encoding/json"
	"fmt"
)

type FetchMaper struct{}

func (fm *FetchMaper) FetchRequestToUnidataFLSource(req dto.FetchRequest) (*source.UnidataFLSource, error) {
	source, err := source.NewUnidataFLSource(req.Address, req.Method, 3, req.Data)
	if err != nil {
		return nil, err
	}
	return source, nil
}

func (fm *FetchMaper) FetchRequestToUnidataULSource(req dto.FetchRequest) (*source.UnidataULSource, error) {
	source, err := source.NewUnidataULSource(req.Address, req.Method, 3, req.Data)
	if err != nil {
		return nil, err
	}
	return source, nil
}

func (fm *FetchMaper) CitizenToFetchResponse(data source.ServiceData) (*dto.FetchResponce, error) {
	citizen, ok := data.(source.Citizen)
	if !ok {
		return nil, fmt.Errorf("ошибка: ожидался тип source.Citizen, но получен %T", data)
	}

	marshalData, err := json.Marshal(&citizen)
	if err != nil {
		return nil, err
	}
	resp := dto.FetchResponce{
		Status:  "Success",
		Message: string(marshalData),
	}
	return &resp, nil
}
