package signaturevalidation

import (
	"bytes"
	"crypto/hmac"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HBOCodeLabs/hurley-kit/llog"
)

// HurleyAuthSignature represents the signed request signature with signed body and validation timestamp
type HurleyAuthSignature struct {
	signedBody          string
	validationTimestamp int64
}

const (
	// SignatureHeaderName is a header name for signed request signature
	SignatureHeaderName = "x-hurley-auth-sign"
	timestampHeaderKey  = "timestamp"
	signatureHeaderKey  = "signature"
	encryptTolerance    = 60000
)

// The example of signature header is "x-hurley-auth-sign: signature=b468fdf5ed96aa8b6cf81ad7cf99bd31b48bff61b37f10b54d1ac7550c79014a;timestamp=1583189404060"
func parseSignatureHeader(header string) (HurleyAuthSignature, error) {
	dataParts := strings.Split(header, ";")
	var authSignature = HurleyAuthSignature{}
	for _, part := range dataParts {
		data := strings.Split(part, "=")
		if len(data) != 2 {
			err := errors.New("parse signature header error")
			return HurleyAuthSignature{}, err
		}
		if data[0] == timestampHeaderKey {
			timestamp, err := strconv.ParseInt(data[1], 10, 64)
			if err != nil {
				return HurleyAuthSignature{}, err
			}
			authSignature.validationTimestamp = timestamp
		} else if data[0] == signatureHeaderKey {
			authSignature.signedBody = data[1]
		}
	}
	if authSignature.signedBody == "" {
		return HurleyAuthSignature{}, errors.New("no header signature found")
	}
	return authSignature, nil
}

func ensureHurleySignatureHeaderAndBody(req *http.Request) (string, error) {
	hurleySignatureHeader := req.Header.Get(SignatureHeaderName)
	if hurleySignatureHeader != "" && req.Body != nil {
		return hurleySignatureHeader, nil
	}
	return "", errors.New("no signature header or body found")
}

func safeCompare(reqHeaderSignature HurleyAuthSignature,
	generatedRequestSignature string) error {
	incomingBody := []byte(reqHeaderSignature.signedBody)
	generatedBody := []byte(generatedRequestSignature)
	timestampDelta := time.Now().UnixNano()/1e6 - reqHeaderSignature.validationTimestamp
	expiredTimestamp := timestampDelta > encryptTolerance
	secureComparison := hmac.Equal(generatedBody, incomingBody)
	if expiredTimestamp || !secureComparison {
		return errors.New(fmt.Sprintf("unable to validate request signature, results: expiredTimestamp: %v; secureComparison: %v", expiredTimestamp, secureComparison))
	} else {
		return nil
	}
}

func validateSignatureHeader(secret string, req *http.Request) error {
	ctxLog := llog.WithCtx(req.Context()).With("method", "validateSignatureHeader")
	hurleySignatureHeader, err := ensureHurleySignatureHeaderAndBody(req)
	if err != nil {
		ctxLog.Error("message", "ensureHurleySignatureHeaderAndBodyError", "error", err)
		return err
	}
	requestSignature, err := parseSignatureHeader(hurleySignatureHeader)
	if err != nil {
		ctxLog.Error("message", "parseSignatureHeaderError", "error", err)
		return err
	}
	validationTimeStamp := strconv.FormatInt(requestSignature.validationTimestamp, 10)
	byts, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ctxLog.Error("message", "convertRequestBodyError", "error", err)
		return errors.New("convert requestBody error")
	}
	// The encoded req.Body will have an empty line added, trim the bytes before validating,
	// otherwise safeCompare(requestSignature, generatedRequestSignature) will always be false
	byts = bytes.TrimRight(byts, "\n")
	generatedRequestSignature, err := generateRequestSignatureHeader(secret, byts, validationTimeStamp)
	if err != nil {
		ctxLog.Error("message", "generateSignatureHeaderError", "error", err)
		return errors.New("generate signature header error")
	}
	err = safeCompare(requestSignature, generatedRequestSignature)
	if err != nil {
		ctxLog.Error("message", "safeCompareFailure", "error", err)
		return err
	}
	return nil
}
