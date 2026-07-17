package agregator

import (
	"context"
	"dstributed-price-monitor/internal/repository"
	"dstributed-price-monitor/internal/source"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type RedisRepo interface {
	Put(ctx context.Context, id string, data []byte) error
	Get(ctx context.Context, id string) ([]byte, error)
	Delete(ctx context.Context, id string) error
}

type Repository interface {
	Create(ctx context.Context, id string, data []byte) error
	GetByID(ctx context.Context, id string) ([]byte, error)
	Update(ctx context.Context, id string, data []byte) ([]byte, error)
	Delete(ctx context.Context, id string) error
}

type Agregator struct {
	pg  Repository
	rds RedisRepo
	mu  sync.Mutex
}

func New(pgRepo Repository, redisRepo RedisRepo) *Agregator {
	return &Agregator{
		pg:  pgRepo,
		rds: redisRepo,
	}
}

func (a *Agregator) Comparable(ctx context.Context, data chan source.ServiceData) {
	concreteRepo, ok := a.pg.(*repository.Postgres)
	if !ok {
		log.Printf("agregator.Agregator.Comparable: incorrect client type for postgreSQL")
		return
	}
	for res := range data {
		switch d := res.(type) {
		case source.Citizen:
			select {
			case <-ctx.Done():
				log.Print("agregator.Agregator.Comparable. context cancel")
				return
			default:
			}
			currentId := d.Data.MdmId
			currentData, err := d.Marshal(ctx)
			if err != nil {
				log.Print(err)
				continue
			}
			isDiffInRedis, err := a.checkDifferenceInRedis(ctx, currentId, currentData)
			if err != nil {
				if strings.Contains(err.Error(), redis.Nil.Error()) {
					data, err := concreteRepo.GetByID(ctx, currentId)
					if err != nil {
						if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) && len(data) == 0 {
							if err := a.writeNewDataInDBAndRedis(ctx, currentId, currentData); err != nil {
								log.Print(err)
							}
							continue
						}
					} else {
						if err := a.writeUpdateDataInDBAndRedis(ctx, currentId, currentData); err != nil {
							log.Print(err)
							continue
						}
					}
				}
				log.Print(err)
				continue
			}
			if isDiffInRedis {
				if err := a.writeUpdateDataInDBAndRedis(ctx, currentId, currentData); err != nil {
					log.Print(err)
					continue
				}
			} else {
				continue
			}
		}
	}
}

func (a *Agregator) checkDifferenceInRedis(ctx context.Context, id string, data []byte) (bool, error) {
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("agregator.Agregator.checkDifferenceInRedis: context cancel")
	default:
	}
	concreteRepo, ok := a.rds.(*repository.Redis)
	if !ok {
		return false, fmt.Errorf("agregator.Agregator.checkDifferenceInRedis: incorrect client type for Redis")
	}

	receivData, err := concreteRepo.Get(ctx, id)
	if err != nil {
		return false, err
	}

	if string(receivData) == string(data) {
		log.Printf("agregator.Agregator.checkDifferenceInRedis: no differences data by id=%s", id)
		log.Printf("agregator.Agregator.checkDifferenceInRedis: to:= %s, after:=%s", string(receivData), string(data))
		return false, nil
	}
	log.Printf("agregator.Agregator.checkDifferenceInRedis: to be differences data by id=%s", id)
	log.Printf("agregator.Agregator.checkDifferenceInRedis: to:= %s, posle:=%s", string(receivData), string(data))
	return true, nil
}

func (a *Agregator) writeNewDataInDBAndRedis(ctx context.Context, id string, data []byte) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("agregator.Agregator.writingDataInDBAndRedis: context cancel")
	default:
	}

	if err := a.pg.Create(ctx, id, data); err != nil {
		return err
	}

	if err := a.rds.Put(ctx, id, data); err != nil {
		return err
	}
	return nil
}

func (a *Agregator) writeUpdateDataInDBAndRedis(ctx context.Context, id string, data []byte) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("agregator.Agregator.writingDataInDBAndRedis: context cancel")
	default:
	}

	if _, err := a.pg.Update(ctx, id, data); err != nil {
		return err
	}

	if err := a.rds.Put(ctx, id, data); err != nil {
		return err
	}

	return nil
}
