package constants

import "time"

// Redis Key Prefixes
const (
	RedisKeyPrefixUser    = "user:"
	RedisKeyPrefixSession = "session:"
	RedisKeyPrefixToken   = "token:"
	RedisKeyPrefixCache   = "cache:"
)

// Redis TTL (Time To Live)
const (
	RedisTTLShort  = 5 * time.Minute
	RedisTTLMedium = 30 * time.Minute
	RedisTTLLong   = 24 * time.Hour
	RedisTTLWeek   = 7 * 24 * time.Hour
)
