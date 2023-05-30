package main

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	cli *redis.Client
}

func (c *RedisClient) InitClient(ctx context.Context, address, password string) error {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	// test connection
	if err := r.Ping(ctx).Err(); err != nil {
		return err
	}

	c.cli = r
	return nil
}
