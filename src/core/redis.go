package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var ctx = context.Background()

const (
	// URL_ID_KEY is global counter
	URL_ID_KEY = "next:url:id"

	// SHORT_LINK_URL_KEY is mapping the short-link to the url
	SHORT_LINK_URL_KEY = "short_link:%s:url"

	// SHORT_LINK_DETAIL_KEY is mapping the short-link to the detail of url
	SHORT_LINK_DETAIL_KEY = "short_link:%s:detail"

	// URL_HASH_KEY is mapping the hash of the url to the short-link
	URL_HASH_KEY = "url_hash:%s:url"
)

// RedisClient contains a redis client
type RedisClient struct {
	Client *redis.Client
}

// URLDetail contains the detail of the short-link
type URLDetail struct {
	URL                 string        `json:"url"`
	CreateAt            string        `json:"create_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

// NewRedisClient create a redis client
func NewRedisClient(addr string, passwd string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db})
	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	return &RedisClient{Client: client}
}
