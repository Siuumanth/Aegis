package shared

import (
	"errors"

	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrBackend        = errors.New("backend error")
	ErrInvalidCommand = errors.New("invalid command")
	ErrGoRedisNil     = goredis.Nil
)
