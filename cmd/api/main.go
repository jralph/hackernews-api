package main

import (
	"flag"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/jralph/hackernews-api/internal/server"
	"github.com/jralph/hackernews-api/pkg/storage"
)

func main() {
	redisHost := flag.String("redis-host", "127.0.0.1:6379", "set the redis host in format of <host>:<port>")

	flag.Parse()

	store := storage.NewRedisStore(
		storage.WithRedisOptions(&redis.Options{
			Addr: *redisHost,
		}),
	)

	svr := server.CreateServer(
		server.WithStorage(store),
	)

	http.ListenAndServe(":8901", svr)
}
