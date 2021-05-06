package redis

import (
	"context"
	"fmt"
	"time"
)

const (
	// TTLKeyNotFound is redis return status to TTL command that simbolize a key not found
	TTLKeyNotFound = -2
	// KeyWithoutTTL is redis return status to TTL command that simbolize a key without TTL set
	KeyWithoutTTL = -1
)

// Redis interface define wich redis methods will be used by leaderboard module
type Redis interface {
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) error
	ExpireAt(ctx context.Context, key string, time time.Time) error
	Ping(ctx context.Context) (string, error)
	SAdd(ctx context.Context, key, member string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key, member string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	ZAdd(ctx context.Context, key string, members ...*Member) error
	ZCard(ctx context.Context, key string) (int64, error)
	ZIncrBy(ctx context.Context, key, member string, increment float64) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]*Member, error)
	ZRank(ctx context.Context, key, member string) (int64, error)
	ZRem(ctx context.Context, key string, members ...string) error
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]*Member, error)
	ZRevRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error)
	ZRevRank(ctx context.Context, key, member string) (int64, error)
	ZScore(ctx context.Context, key, member string) (float64, error)
}

// Member is a struct to be used by sorted set range operations
type Member struct {
	Member string
	Score  float64
}

func (m *Member) String() string {
	return fmt.Sprintf("{Member: %s, Score: %f}", m.Member, m.Score)
}
