package idempotency

import (
	"encoding/json"
	"time"

	cacheRepo "github.com/faisalhardin/medilink/internal/entity/repo/cache"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
)

const (
	defaultTTL      = 5 // seconds
	defaultPollWait = 100 * time.Millisecond
	defaultPollMax  = 2 * time.Second

	statusPending   = "pending"
	statusCompleted = "completed"
)

type record struct {
	Status      string          `json:"status"`
	RequestHash string          `json:"request_hash"`
	Response    json.RawMessage `json:"response,omitempty"`
}

// Service provides cache-backed request idempotency (claim, poll, complete, release).
type Service struct {
	cache    cacheRepo.Caching
	ttl      int
	pollWait time.Duration
	pollMax  time.Duration
}

type Option func(*Service)

func WithTTL(seconds int) Option {
	return func(s *Service) {
		if seconds > 0 {
			s.ttl = seconds
		}
	}
}

func WithPoll(wait, max time.Duration) Option {
	return func(s *Service) {
		if wait > 0 {
			s.pollWait = wait
		}
		if max > 0 {
			s.pollMax = max
		}
	}
}

func New(cache cacheRepo.Caching, opts ...Option) *Service {
	s := &Service{
		cache:    cache,
		ttl:      defaultTTL,
		pollWait: defaultPollWait,
		pollMax:  defaultPollMax,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Acquire claims the idempotency key for requestHash.
//
//	replayed=true  → return the cached completed response; do not run the operation
//	replayed=false → key claimed; run the operation, then Complete or Release
func Acquire[T any](s *Service, key, requestHash string) (response T, replayed bool, err error) {
	var zero T
	if s == nil || s.cache == nil {
		return zero, false, nil
	}

	cached, getErr := s.cache.Get(key)
	if getErr == nil {
		var rec record
		if jsonErr := json.Unmarshal([]byte(cached), &rec); jsonErr == nil {
			if rec.RequestHash != requestHash {
				err = commonerr.SetNewUnprocessableEntityError(
					"idempotency_key_conflict",
					"idempotency key was already used with a different request payload",
				)
				return zero, false, err
			}
			if rec.Status == statusCompleted {
				if len(rec.Response) > 0 {
					if jsonErr := json.Unmarshal(rec.Response, &response); jsonErr != nil {
						return zero, false, jsonErr
					}
				}
				return response, true, nil
			}
			// pending — another concurrent request is in flight
			return pollForCompletion[T](s, key)
		}
	}

	pendingJSON, _ := json.Marshal(record{
		Status:      statusPending,
		RequestHash: requestHash,
	})

	claimed, setErr := s.cache.SetNX(key, string(pendingJSON), s.ttl)
	if setErr != nil {
		// cache error is non-fatal; proceed without protection
		return zero, false, nil
	}
	if !claimed {
		return pollForCompletion[T](s, key)
	}

	return zero, false, nil
}

// Complete stores the successful response under the idempotency key.
func Complete[T any](s *Service, key, requestHash string, response T) {
	if s == nil || s.cache == nil {
		return
	}
	respJSON, _ := json.Marshal(response)
	completedJSON, _ := json.Marshal(record{
		Status:      statusCompleted,
		RequestHash: requestHash,
		Response:    respJSON,
	})
	s.cache.SetWithExpire(key, string(completedJSON), s.ttl)
}

// Release deletes the idempotency key so the client can retry with the same key.
func Release(s *Service, key string) {
	if s == nil || s.cache == nil {
		return
	}
	s.cache.Del(key)
}

func pollForCompletion[T any](s *Service, key string) (response T, replayed bool, err error) {
	var zero T
	deadline := time.Now().Add(s.pollMax)
	for time.Now().Before(deadline) {
		time.Sleep(s.pollWait)
		cached, getErr := s.cache.Get(key)
		if getErr != nil {
			break
		}
		var rec record
		if jsonErr := json.Unmarshal([]byte(cached), &rec); jsonErr != nil {
			break
		}
		if rec.Status == statusCompleted {
			if len(rec.Response) > 0 {
				if jsonErr := json.Unmarshal(rec.Response, &response); jsonErr != nil {
					return zero, false, jsonErr
				}
			}
			return response, true, nil
		}
	}
	return zero, false, commonerr.SetNewIdempotencyInProgressError()
}
