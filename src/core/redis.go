package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mattheath/base62"
	"github.com/speps/go-hashids"
	"time"
)

var ctx = context.Background()

const (
	// UrlIdKey is global counter
	UrlIdKey = "next:url:id"

	// ShortLinkUrlKey is mapping the short-link to the url
	ShortLinkUrlKey = "short_link:%s:url"

	// ShortLinkDetailKey is mapping the short-link to the detail of url
	ShortLinkDetailKey = "short_link:%s:detail"

	// UrlHashKey is mapping the hash of the url to the short-link
	UrlHashKey = "url_hash:%s:url"
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
	h := toHash(url)

	// fetch it if the url is cached
	d, err := r.Client.Get(ctx, fmt.Sprintf(UrlHashKey, h)).Result()
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
	err = r.Client.Incr(ctx, UrlIdKey).Err()
	if err != nil {
		return "", err
	}

	// encode global counter to base62
	id, err := r.Client.Get(ctx, UrlIdKey).Int64()
	if err != nil {
		return "", err
	}
	eid := base62.EncodeInt64(id)

	// store the url against this encoded id
	err = r.Client.Set(ctx, fmt.Sprintf(ShortLinkUrlKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// store the url against the hash of it
	err = r.Client.Set(ctx, fmt.Sprintf(UrlHashKey, h), eid, time.Minute*time.Duration(exp)).Err()
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
	err = r.Client.Set(ctx, fmt.Sprintf(ShortLinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	return eid, nil
}

// ShortlinkInfo returns the detail of the shortlink
func (r *RedisClient) ShortlinkInfo(eid string) (interface{}, error) {
	d, err := r.Client.Get(ctx, fmt.Sprintf(ShortLinkDetailKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, errors.New("unknown short-link")}
	} else if err != nil {
		return "", err
	} else {
		return d, nil
	}
}

// Unshorten convent short-link to url
func (r *RedisClient) Unshorten(eid string) (string, error) {
	url, err := r.Client.Get(ctx, fmt.Sprintf(ShortLinkUrlKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, errors.New(fmt.Sprintf("%s short-link expired", eid))}
	} else if err != nil {
		return "", err
	} else {
		return url, nil
	}
}

func toHash(url string) interface{} {
	hd := hashids.NewData()
	hd.Salt = url
	hd.MinLength = 0
	h, _ := hashids.NewWithData(hd)
	r, _ := h.Encode([]int{45, 434, 1313, 99})
	return r
}
