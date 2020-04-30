package signaturevalidation

import (
	"crypto"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const requestHeaderFormat = "signature=%s;timestamp=%s"

// Creates a header object with a signed payload
func CreateSignedRequestHeader(secret string, payload interface{}) (string, error) {
	currentTime := time.Now().UnixNano() / 1e6
	timestamp := strconv.FormatInt(currentTime, 10)
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	header, err := generateRequestSignatureHeader(secret, body, timestamp)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(requestHeaderFormat, header, timestamp), nil
}

// Accepts a secret and a data payload, returns a SHA256 encoded header
func generateRequestSignatureHeader(secret string, body []byte, currentTime string) (string, error) {
	h := hmac.New(crypto.SHA256.New, []byte(secret))
	h.Write([]byte(currentTime))
	h.Write([]byte([]byte("_")))
	h.Write(body)
	sha := hex.EncodeToString(h.Sum(nil))
	return sha, nil
}
