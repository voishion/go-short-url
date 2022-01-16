package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mattheath/base62"
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

// Shorten convert url to shortlink
func (r *RedisClient) Shorten(url string, exp int64) (string, error) {
	// convent url to sha1 hash
	h := toSha1(url)

	// fetch it if the url is cached
	d, err := r.Client.Get(ctx, fmt.Sprintf(URL_HASH_KEY, h)).Result()
	if err == redis.Nil {
		// not existed, nothing to do
	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
			// expiration, nothing to do
		} else {
			return d, nil
		}
	}

	// increase the global counter
	err = r.Client.Incr(ctx, URL_ID_KEY).Err()
	if err != nil {
		return "", err
	}

	// encode global counter to base62
	id, err := r.Client.Get(ctx, URL_ID_KEY).Int64()
	if err != nil {
		return "", err
	}
	eid := base62.EncodeInt64(id)

	// store the url against this encoded id
	err = r.Client.Set(ctx, fmt.Sprintf(SHORT_LINK_URL_KEY, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// store the url against the hash of it
	err = r.Client.Set(ctx, fmt.Sprintf(URL_HASH_KEY, h), eid, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(&URLDetail{
		URL:                 url,
		CreateAt:            time.Now().String(),
		ExpirationInMinutes: time.Duration(exp)})
	if err != nil {
		return "", err
	}

	// store the url detail against this encoded id
	err = r.Client.Set(ctx, fmt.Sprintf(SHORT_LINK_DETAIL_KEY, eid), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	return eid, nil
}

func toSha1(url string) interface{} {
	return nil
}
