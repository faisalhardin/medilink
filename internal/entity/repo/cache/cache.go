package cache

type Caching interface {
	Get(key string) (string, error)
	Del(key string) (int64, error)
	SetWithExpire(key string, value interface{}, expire int) (string, error)
	// SetNX sets key to value with TTL only if key does not exist (atomic).
	// Returns true if the key was set, false if it already existed.
	SetNX(key string, value interface{}, expire int) (bool, error)
}
