package config

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(host string, port int, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	return rdb
}
