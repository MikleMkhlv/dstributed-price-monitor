package mapper

import (
	dto "dstributed-price-monitor/api/dto"
	"dstributed-price-monitor/internal/source"
	"log"
)

type MapperShd struct{}

func (mshd *MapperShd) RecordToRequestForFether(rec source.Record) dto.FetchRequest {
	switch v := rec.(type) {
	case *source.UnidataFLSource:
		req := dto.FetchRequest{
			Type:    "unidata_fl",
			Address: v.Url,
			Method:  v.Method,
			Data:    v.MdmIds,
		}
		return req
	case *source.UnidataULSource:
		req := dto.FetchRequest{
			Type:    "unidata_fl",
			Address: v.Url,
			Method:  v.Method,
			Data:    v.MdmIds,
		}
		return req
	default:
		log.Print("mapper.MapperShd.RecordToRequestForFether: unknown source type")
		return dto.FetchRequest{}
	}
}
