package cache

type Caching interface {
	Get(key string) (string, error)
	Del(key string) (int64, error)
	SetWithExpire(key string, value interface{}, expire int) (string, error)
}
