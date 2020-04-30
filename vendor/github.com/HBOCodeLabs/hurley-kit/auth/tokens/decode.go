package tokens

import (
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

const (
	// algortihmHMAC256 is the algorithm used to encode the auth tokens
	algorithmHMAC256 = "HS256"

	// typeJSONWebToken is the string we expect to find in the "typ" header of a
	// valid auth token
	typeJSONWebToken = "JWT"
)

// A Decoder reads and decodes HBO GO auth tokens, using a secret to verify the
// signature.
type Decoder struct {
	verificationBytes []byte
}

// NewDecoder creates an instance of Decoder that uses the bytes in the
// `secret` string to verify the token's signature.
func NewDecoder(secret string) *Decoder {
	return &Decoder{
		verificationBytes: []byte(secret),
	}
}

// Decode takes an encoded token value (typically a "Bearer" value from an
// "authorization" header), decodes it, and performs signature verification.
func (d *Decoder) Decode(tokenValue string) (Tokener, error) {
	token := &Token{}

	t, err := d.parse(tokenValue, &token.claims, token)
	if err == nil {
		if t.Valid {
			// success!
			return token, nil
		}
		return nil, errors.New("token was not valid after parsing and verification")
	}

	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("invalid token: malformed token value")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return nil, errors.New("invalid token: token is expired or not active yet")
		} else {
			return nil, fmt.Errorf("validation error: could not parse/validate token: %s", err.Error())
		}
	}

	// some other error
	return nil, err
}

// setRawer is the interface that encapsulates the setRaw method, used to set
// the "raw" encoded token value on any type that implements the interface.
type setRawer interface {
	setRaw(string)
}

// parse will perform the token parsing and verification.  The secret used at
// the time of construction of the Decoder will be used to verify the
// signature, in the verification callback func.
func (d *Decoder) parse(tokenValue string, claims jwt.Claims, token setRawer) (*jwt.Token, error) {
	var t *jwt.Token
	var err error

	// perform the token parsing and verification.  The secret used at the time
	// of construction will be used to verify the signature, in the verification
	// callback func.
	t, err = jwt.ParseWithClaims(tokenValue, claims, func(t *jwt.Token) (interface{}, error) {
		// verify that the algorithm and type are what we expect
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			// verify that we're using HS256
			if t.Header["alg"] != algorithmHMAC256 || t.Header["typ"] != typeJSONWebToken {
				return nil, fmt.Errorf("unexpected algorithm or type in token headers: %#v", t.Header)
			}
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		token.setRaw(t.Raw)
		// return the key to use for verification
		// This comes from the values in hurley-keys/hurley-secret s3 buckets, staging/token.[txt|json]
		return d.verificationBytes, nil
	})

	return t, err
}
