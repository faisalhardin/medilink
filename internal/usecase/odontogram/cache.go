package odontogram

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/repo/cache"
)

const (
	snapshotCacheTTL = 5 * time.Minute
	cacheKeyPrefix   = "odontogram:"
)

// CachedSnapshot represents a cached snapshot with metadata
type CachedSnapshot struct {
	Snapshot            model.OdontogramSnapshot `json:"snapshot"`
	MaxLogicalTimestamp int64                    `json:"max_logical_timestamp"`
	MaxSequenceNumber   int64                    `json:"max_sequence_number"`
	LastUpdated         int64                    `json:"last_updated"`
}

// SnapshotCache manages in-memory caching of snapshots
type SnapshotCache struct {
	cache cache.Caching
}

// NewSnapshotCache creates a new snapshot cache
func NewSnapshotCache(cache cache.Caching) *SnapshotCache {
	return &SnapshotCache{cache: cache}
}

// Get retrieves a snapshot from cache
func (sc *SnapshotCache) Get(ctx context.Context, patientUUID string) (*CachedSnapshot, error) {
	key := getCacheKey(patientUUID)

	data, err := sc.cache.Get(key)
	if err != nil {
		return nil, err
	}

	var cached CachedSnapshot
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		return nil, err
	}

	return &cached, nil
}

// Set stores a snapshot in cache
func (sc *SnapshotCache) Set(ctx context.Context, patientUUID string, snapshot model.OdontogramSnapshot, maxLogicalTimestamp, maxSequenceNumber, lastUpdated int64) error {
	key := getCacheKey(patientUUID)

	cached := CachedSnapshot{
		Snapshot:            snapshot,
		MaxLogicalTimestamp: maxLogicalTimestamp,
		MaxSequenceNumber:   maxSequenceNumber,
		LastUpdated:         lastUpdated,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	_, err = sc.cache.SetWithExpire(key, string(data), int(snapshotCacheTTL.Seconds()))
	if err != nil {
		return err
	}

	return nil
}

// Invalidate removes a snapshot from cache
func (sc *SnapshotCache) Invalidate(ctx context.Context, patientUUID string) error {
	key := getCacheKey(patientUUID)
	_, err := sc.cache.Del(key)
	if err != nil {
		return err
	}

	return nil
}

// getCacheKey generates a cache key for a patient
func getCacheKey(patientUUID string) string {
	return fmt.Sprintf("%s%s", cacheKeyPrefix, patientUUID)
}
