package secrets

import (
	"errors"
	"time"

	"github.com/HBOCodeLabs/hurley-kit/llog"
)

// NewStore creates and returns an instance of a Store
//  ttl indicates how long secret store should cache the secret in memory
//  bucket is name of s3 bucket where secrets are held
//  region is the AWS region where secrets are held
func NewStore(ttl time.Duration, bucket, region string) (*Store, error) {
	if ttl <= zeroDuration {
		return nil, errors.New("must specify a ttl greater than 0")
	}

	if bucket == "" {
		return nil, errors.New("must specify bucket")
	}

	if region == "" {
		return nil, errors.New("must specify region")
	}

	return &Store{
		cacheEntries:   make(map[string]cacheEntry),
		ttl:            ttl,
		region:         region,
		bucket:         bucket,
		s3ObjectGetter: &s3Client{},
		utcNowGetter:   &clock{},
	}, nil
}

// Store is the struct used to access hurley secrets
type Store struct {
	cacheEntries   map[string]cacheEntry
	ttl            time.Duration
	region         string
	bucket         string
	s3ObjectGetter s3ObjectGetter
	utcNowGetter   utcNowGetter
}

// Get returns the secret by key
func (s *Store) Get(key string) ([]byte, error) {
	entry, ok := s.cacheEntries[key]
	llog.Debug("event", "resultFromMap", "key", key, "ok", ok)

	utcNow := s.utcNowGetter.getUTCNow()
	if ok && entry.expires.After(*utcNow) {
		llog.Debug("event", "returnFromCache", "key", key)
		return entry.byts, nil
	}

	byts, err := s.refresh(key)

	if err != nil {
		llog.Error("event", "failedToGetSecret", "err", err)
		return nil, err
	}

	return byts, nil
}

// refresh will reload the secret from S3 via s3ObjectGetter and cache
// successful result in memory
func (s *Store) refresh(key string) ([]byte, error) {
	// TODO: multiplexer
	byts, err := s.s3ObjectGetter.getObject(s.region, s.bucket, key)

	if err != nil {
		return nil, err
	}

	utcNow := s.utcNowGetter.getUTCNow()
	expires := (*utcNow).Add(s.ttl)
	llog.Debug("event", "storingCacheEntry", "key", key, "expires", expires)
	s.cacheEntries[key] = cacheEntry{
		expires: expires,
		byts:    byts,
	}

	return byts, nil
}
