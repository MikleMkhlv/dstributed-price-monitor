package source

import (
	"context"
	"dstributed-price-monitor/config"
	"log"
)

type Record interface {
	Pool(ctx context.Context, outCh chan ServiceData, errCh chan error)
}

type ServiceData interface {
	IsImpl()
}

type ServiceAgregats interface {
	Comparable(ctx context.Context, data chan ServiceData)
}
type Source struct {
	Sources map[string]Record
}

func NewSource(cfg config.Config) *Source {
	sources := make(map[string]Record)
	for _, src := range cfg.Sources {
		switch src.Type {
		case "unidata_fl":
			data, ok := src.Data.(config.UnidataFLUL)
			if !ok {
				log.Printf("source.NewSource: unexpected data type for %s", src.Type)
				continue
			}
			s, err := NewUnidataFLSource(src.URL, src.Method, cfg.Scheduler.Timeout, data.MdmIds)
			if err != nil {
				log.Printf("source.NewSource: %v\n", err)
				continue
			}
			sources[src.Type] = s
		case "unidata_ul":
			data, ok := src.Data.(config.UnidataFLUL)
			if !ok {
				log.Printf("source.NewSource: unexpected data type for %s", src.Type)
				continue
			}
			s, err := NewUnidataULSource(src.URL, src.Method, cfg.Scheduler.Timeout, data.MdmIds)
			if err != nil {
				log.Printf("source.NewSource: %v\n", err)
				continue
			}
			sources[src.Type] = s
		default:
			log.Printf("source.NewSource: unknown type source. type = %s\n", src.Type)
			continue

		}
	}
	log.Print("source.NewSource: create sources complete: ")
	// for k := range sources {
	// 	log.Print(k + ",")
	// }
	return &Source{
		Sources: sources,
	}
}
