package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/jralph/hackernews-api/internal/scraper"
)

var ctx = context.Background()

type Redis struct {
	client *redis.Client
}

type Option func(*Redis)

func WithRedisOptions(opts *redis.Options) Option {
	return func(r *Redis) {
		r.client = redis.NewClient(opts)
	}
}

func WithRedis(client *redis.Client) Option {
	return func(r *Redis) {
		r.client = client
	}
}

func NewRedisStore(opts ...Option) *Redis {
	client := &Redis{}

	for _, opt := range opts {
		opt(client)
	}

	if client.client == nil {
		panic(fmt.Errorf("storage: must provide either `WithRedis` or `WithRedisOptions` to NewRedisStore"))
	}

	return client
}

func (r *Redis) SaveTopStories(topStories scraper.TopStoriesResponse) error {
	data, _ := json.Marshal(topStories)
	return r.client.Set(ctx, "hn_top_stories", data, 0).Err()
}

func (r *Redis) SaveItem(item *scraper.ItemResponse) error {
	data, _ := json.Marshal(item)
	return r.client.Set(ctx, fmt.Sprintf("hn_item_%s_%d", item.Type, item.ID), data, 0).Err()
}

func (r *Redis) DeleteItem(item *scraper.ItemResponse) error {
	return r.client.Del(ctx, fmt.Sprintf("hn_item_%s_%d", item.Type, item.ID)).Err()
}

func (r *Redis) GetAllItems() ([]int, error) {
	keys, _ := r.client.Keys(ctx, "hn_item_*").Result()

	matchesPost := regexp.MustCompile(`hn_item_(story|job|poll|comment|pollopt)_([0-9]+)`)

	var items []int
	for _, key := range keys {
		if !matchesPost.MatchString(key) {
			continue
		}

		match := matchesPost.FindStringSubmatch(key)

		foundId, _ := strconv.Atoi(match[2])
		items = append(items, foundId)
	}

	return items, nil
}

func (r *Redis) GetAllPosts(postType *string) ([]int, error) {
	keys, _ := r.client.Keys(ctx, "hn_item_*").Result()

	searchTypes := "story|job|poll"
	if postType != nil {
		searchTypes = *postType
	}

	matchesPost := regexp.MustCompile(fmt.Sprintf("hn_item_(%s)_([0-9]+)", searchTypes))

	var items []int
	for _, key := range keys {
		if !matchesPost.MatchString(key) {
			continue
		}

		match := matchesPost.FindStringSubmatch(key)

		foundId, _ := strconv.Atoi(match[2])
		items = append(items, foundId)
	}

	return items, nil
}

func (r *Redis) GetItem(id int) (*scraper.ItemResponse, error) {
	keys, _ := r.client.Keys(ctx, fmt.Sprintf("hn_item_*_%d", id)).Result()

	if len(keys) == 0 {
		return nil, nil
	}

	data, _ := r.client.Get(ctx, keys[0]).Result()

	var scrapedItem scraper.ItemResponse

	err := json.Unmarshal([]byte(data), &scrapedItem)

	return &scrapedItem, err
}
