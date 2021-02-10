package storage

import (
	"testing"

	"github.com/jralph/hackernews-api/internal/server"

	"github.com/go-redis/redis/v8"

	"github.com/jralph/hackernews-api/internal/scraper"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewRedisStore(
		WithRedisOptions(&redis.Options{}),
	)

	t.Run("NewRedisStore returns instance of Redis and implements scraper saver interface", func(t *testing.T) {
		_, okSaver := interface{}(client).(scraper.Saver)
		_, okStorage := interface{}(client).(server.Storage)
		require.IsType(t, &Redis{}, client)
		require.True(t, okSaver)
		require.True(t, okStorage)
	})
}
