package secrets

import (
	"time"
)

const zeroDuration = time.Duration(0)

type cacheEntry struct {
	expires time.Time
	byts    []byte
}
