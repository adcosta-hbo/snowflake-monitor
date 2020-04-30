package strutil

import (
	"math/rand"
	"sync"
	"time"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	mutex      sync.Mutex // used to synchronize access to the above source
)

const (
	// possible letters to choose from for the random string
	letterBytes   = "0123456789abcdef"
	letterIdxBits = 4                    // 4 bits to represent a letter index, using the above letter choices
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandomHexString generates a new pseudo-random string, consisting of `len` lowercase
// hex digits.  The function is safe for use in multiple concurrent goroutines.
func RandomHexString(len int) string {
	return randStringBytesMask(len)
}

// randStringBytesMask will return a pseudo-random string of lowercase hex
// digits of length `n`.
//
// The basic algortihm is ~5x faster than the naive solution (choosing n random
// indices in a slice of letter runes), and comes from the excellent SO post at:
// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func randStringBytesMask(n int) string {
	b := make([]byte, n)
	// A randSource.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, atomicInt63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = atomicInt63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func atomicInt63() int64 {
	mutex.Lock()
	v := randSource.Int63()
	mutex.Unlock()
	return v
}
