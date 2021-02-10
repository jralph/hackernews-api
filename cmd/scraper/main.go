package main

import (
	"flag"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/jralph/hackernews-api/internal/scraper"
	"github.com/jralph/hackernews-api/pkg/hnclient"
	"github.com/jralph/hackernews-api/pkg/storage"
)

func main() {
	redisHost := flag.String("redis-host", "127.0.0.1:6379", "set the redis host in format of <host>:<port>")
	workers := flag.Int("workers", 100, "set the number of works to run when scraping content")

	flag.Parse()

	saver := storage.NewRedisStore(
		storage.WithRedisOptions(&redis.Options{
			Addr: *redisHost,
		}),
	)
	client := hnclient.NewClient()
	s := scraper.NewScraper(
		scraper.WithSaver(saver),
		scraper.WithClient(client),
		scraper.WithWorkerCount(*workers),
	)

	items, err := s.Scrape()
	if err != nil {
		panic(fmt.Errorf("scraper: error running scrape: %s", err))
	}

	fmt.Printf("scraper: successfully scraped %d items\n", items)
}
