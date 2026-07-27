package mapper

import (
	dto "dstributed-price-monitor/api/dto"
	"dstributed-price-monitor/internal/source"
	"encoding/json"
	"fmt"
	"log"
)

type FetchMaper struct{}

func (fm *FetchMaper) FetchRequestToUnidataFLSource(req dto.FetchRequest) (*source.UnidataFLSource, error) {
	log.Printf("mapper.FetchMaper.FetchRequestToUnidataFLSource(DEBUG): %v", req)
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
	switch d := data.(type) {
	case source.Citizen:
		marshalData, err := json.Marshal(&d)
		if err != nil {
			return nil, err
		}
		resp := dto.FetchResponce{
			Status:  "Success",
			Message: string(marshalData),
		}
		return &resp, nil
	case source.Organization:
		marshalData, err := json.Marshal(&d)
		if err != nil {
			return nil, err
		}
		resp := dto.FetchResponce{
			Status:  "Success",
			Message: string(marshalData),
		}
		return &resp, nil
	default:
		log.Print("mapper.FetchMaper.CitizenToFetchResponse: uncnown type")
		return nil, fmt.Errorf("Uncnovn type for mapperFetc")
	}
}
