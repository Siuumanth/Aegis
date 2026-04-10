package redis

import (
	"errors"
	"net"

	"github.com/redis/go-redis/v9"
	gobreaker "github.com/sony/gobreaker/v2"
)

func isRedisConnError(err error) bool {
	if err == nil || err == redis.Nil || err == gobreaker.ErrOpenState {
		return false
	}
	// only network/conn errors should trip CB
	var netErr net.Error
	return errors.As(err, &netErr)
}
