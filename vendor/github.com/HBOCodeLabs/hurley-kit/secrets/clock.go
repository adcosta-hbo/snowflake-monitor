package secrets

import "time"

// utcNowGetter is an interface to return current utc time
type utcNowGetter interface {
	getUTCNow() *time.Time
}

// clock is a struct implementing utcGetter and returns current utc time
type clock struct{}

func (c *clock) getUTCNow() *time.Time {
	utc := time.Now().UTC()
	return &utc
}
