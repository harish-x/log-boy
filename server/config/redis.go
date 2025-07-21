package config

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	s := rdb.Ping(context.Background())
	if s.Err() != nil {
		return nil, s.Err()
	}
	return rdb, nil
}
