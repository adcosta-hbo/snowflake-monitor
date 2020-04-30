/**
 * @preserve Copyright (c) 2017 Home Box Office, Inc. as an unpublished
 * work. Neither this material nor any portion hereof may be copied or
 * distributed without the express written consent of Home Box Office, Inc.
 *
 * This material also contains proprietary and confidential information
 * of Home Box Office, Inc. and its suppliers, and may not be used by or
 * disclosed to any person, in whole or in part, without the prior written
 * consent of Home Box Office, Inc.
 */

package llog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// registerLogfmtEncoder registers the Logfmt encoder.
func registerLogfmtEncoder() error {
	return zap.RegisterEncoder("logfmt",
		func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
			return NewLogfmtEncoder(cfg), nil
		})
}

// newLogfmtEncoderConfig returns an opinionated EncoderConfig for logfmt
func newLogfmtEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "ts",
		LevelKey:       "level",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // ie. level=INFO
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // ie. 2017-08-03T15:33:11.078-0700
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // ie. line="main.go:43" if enabled
	}
}
