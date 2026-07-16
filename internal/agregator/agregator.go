package agregator

import "context"

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
