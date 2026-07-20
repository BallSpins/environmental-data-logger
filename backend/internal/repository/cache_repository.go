package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	rdb *redis.Client
}

func NewCacheRepository(rdb *redis.Client) *CacheRepository {
	return &CacheRepository{rdb: rdb}
}

// BufferSensorData memasukkan payload biner langsung ke dalam Redis List (Buffer FIFO) tanpa alokasi JSON
func (r *CacheRepository) BufferSensorData(ctx context.Context, payload []byte) error {
	key := "queue:sensor_ingestion"

	// LPUSH ke Redis List (Sangat cepat karena berjalan sepenuhnya di RAM, tanpa overhead JSON marshalling)
	return r.rdb.LPush(ctx, key, payload).Err()
}
