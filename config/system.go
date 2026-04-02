package config

func applyServerDefaults(s *ServerConfig) {
	if s.Host == "" {
		s.Host = DefaultServerHost
	}
	if s.Port == 0 {
		s.Port = DefaultServerPort
	}
	if s.ReadTimeout == 0 {
		s.ReadTimeout = DefaultServerReadTimeout
	}
	if s.WriteTimeout == 0 {
		s.WriteTimeout = DefaultServerWriteTimeout
	}
}

func applyRedisDefaults(r *RedisConfig) {
	if r.Address == "" {
		r.Address = DefaultRedisAddress
	}
	if r.PoolSize == 0 {
		r.PoolSize = DefaultRedisPoolSize
	}
	if r.MinIdleConns == 0 {
		r.MinIdleConns = DefaultRedisMinIdleConns
	}
	if r.DialTimeout == 0 {
		r.DialTimeout = DefaultRedisDialTimeout
	}
	if r.ReadTimeout == 0 {
		r.ReadTimeout = DefaultRedisReadTimeout
	}
	if r.WriteTimeout == 0 {
		r.WriteTimeout = DefaultRedisWriteTimeout
	}
	if r.MaxRetries == 0 {
		r.MaxRetries = DefaultRedisMaxRetries
	}
}
