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
	"context"
	"io"
	"os"
	"time"

	"github.com/HBOCodeLabs/hurley-kit/contextdefs"
	abws "github.com/HBOCodeLabs/hurley-kit/llog/buffered"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level represents log levels which are sorted based on the underlying number associated to it.
type Level uint8

// Definitions of available levels: Debug < Info < Warning < Error.
const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
)

// These are the names that'll appear in the log in hurley. ie (traceId=abc123) They are case sensitive
const (
	uberTraceID = "uberTraceID"
	keyTraceID  = "traceId"
	keySpanID   = "spanId"
	keyParentID = "parentId"
	// "caller" is a convention we've been using in Hurley.  To keep the variable name consistent, I call it
	// keyCaller
	keyCaller = "caller"
	keyStack  = "stack"

	maxBufferSize = 1024 * 32              // 32k
	maxDelay      = time.Millisecond * 100 // 100ms https://github.com/HBOCodeLabs/Hurley-Common/blob/master/lib/util/AsyncStdoutLogger.js#L17
)

// globalLogger is the global logging instance
var globalLogger *Logger

// InitWith is a method to initialize all logging with certain values.
// Call this method in e.g. main.go
func InitWith(keyvals ...interface{}) {
	// Ensure old logger is flushed before initialising a new one
	globalLogger.Sync()
	globalLogger = globalLogger.With(keyvals...)
	globalLogger.Sync()
}

// Logger represents an instantiable logger. It's a wrapper for zap's SugarLogger.
// The Level is kept so that it can be changed later.  Note that logging can be performed
// on a logging instance directly, or the global functions such as Error, Info,
// etc. can be used
type Logger struct {
	zaplevel zap.AtomicLevel // Need to keep the level to set it later
	logger   *zap.SugaredLogger
}

// NewLogger sets up a wrapper of zap's sugar logger.
// For example,
//  myLogger := llog.NewLogger(os.Stdout, INFO)
func NewLogger(out io.Writer, l Level) *Logger {
	zapLevel := zap.NewAtomicLevelAt(getZapLevel(l))
	baseLogger := zap.New(
		zapcore.NewCore(
			NewLogfmtEncoder(newLogfmtEncoderConfig()),
			zapcore.Lock(zapcore.AddSync(out)),
			&zapLevel,
		),
		zap.ErrorOutput(abws.NewAsyncBufferWriteSyncer(os.Stderr, maxBufferSize, maxDelay)), // internal zap error output.
	)

	sugarLogger := baseLogger.Sugar()
	return &Logger{
		zaplevel: zapLevel,
		logger:   sugarLogger,
	}
}

// Level returns the current log level of this logger instance
func (l *Logger) Level() Level {
	return getLlogLevel(l.zaplevel.Level())
}

// Info logs a set of keys/values at the INFO level (if enabled)
func (l *Logger) Info(keyvals ...interface{}) {
	// The first argument is meant to go with Zap's MessageKey, which is not used.  So we set it to ""
	l.logger.Infow("", keyvals...)
}

// Fatal logs an error at the ERROR level, then exits the process with an exit code of 1.
func (l *Logger) Fatal(keyvals ...interface{}) {
	keyvals = append(keyvals, zap.Stack(keyStack))
	// The first argument is meant to go with Zap's MessageKey, which is not used.  So we set it to ""
	l.logger.Fatalw("", keyvals...)
}

// Debug logs a set of keys/values at the DEBUG level (if enabled)
func (l *Logger) Debug(keyvals ...interface{}) {
	// The first argument is meant to go with Zap's MessageKey, which is not used.  So we set it to ""
	l.logger.Debugw("", keyvals...)
}

// Warn logs a set of keys/values at the WARNING level (if enabled)
func (l *Logger) Warn(keyvals ...interface{}) {
	// The first argument is meant to go with Zap's MessageKey, which is not used.  So we set it to ""
	l.logger.Warnw("", keyvals...)
}

// Error logs a set of keys/values at the ERROR level (if enabled)
func (l *Logger) Error(keyvals ...interface{}) {
	keyvals = append(keyvals, zap.Stack(keyStack))
	// The first argument is meant to go with Zap's MessageKey, which is not used.  So we set it to ""
	l.logger.Errorw("", keyvals...)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// WithCtx extracts the traceID, spanID, and caller info from the context and adds them to the logging context, and
// returns a copy of the original logger instance with the new context.  The original instance is unchanged.
// For example,
//  var ctx = Context.Background()
//
//  // setup tracing information.  (in a middleware)
// 	ctx = context.WithValue(ctx, contextdefs.TraceID, "testTrace123")
//	ctx = context.WithValue(ctx, contextdefs.SpanID, "testSpan123")
//	ctx = context.WithValue(ctx, contextdefs.HBOCaller, "pickup")
//
//  myLogger := NewLogger(os.Stderr, INFO)
//  ctxLogger := myLogger.WithCtx(ctx)
//  ctxLogger.Info("hello", "world", "username", "dcarney")
//  ctxLogger.Info( "username", "jsong")
//
//  // Result is:  ts="2016-07-18T17:15:42Z", level="INFO", traceId="testTrace123", spanId="testSpan123", caller="pickup", hello="world", username="dcarney"
//  // ts="2016-07-18T17:15:43Z", level="INFO", traceId="testTrace123", spanId="testSpan123", caller="pickup", username="jsong"
func (l *Logger) WithCtx(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	var keyvals []interface{}
	traceID := ctx.Value(contextdefs.TraceID)
	if traceID != nil {
		keyvals = append(keyvals, keyTraceID, traceID.(string))
	}

	spanID := ctx.Value(contextdefs.SpanID)
	if spanID != nil {
		keyvals = append(keyvals, keySpanID, spanID.(string))
	}

	parentID := ctx.Value(contextdefs.ParentSpanID)
	if parentID != nil {
		keyvals = append(keyvals, keyParentID, parentID.(string))
	}

	callerID := ctx.Value(contextdefs.HBOCaller)
	if callerID != nil {
		keyvals = append(keyvals, keyCaller, callerID.(string))
	}

	uberID := ctx.Value(contextdefs.UberHeader)
	if uberID != nil {
		keyvals = append(keyvals, uberTraceID, uberID.(string))
	}

	if len(keyvals) == 0 {
		return l
	}

	return &Logger{
		zaplevel: l.zaplevel,
		logger:   l.logger.With(keyvals...),
	}
}

// With allows static values to be added to the llog, so that every log contains these
// values. The original instance remains unchanged.
// For example,
//  moduleLogger := llog.With("service", "test-service")
//  moduleLogger.Info("hello", "world")
//
//  // Result is:
//  // ts="2016-07-18T17:15:42Z", level="INFO",  service="test-service", hello="world"
// This can also be chained for example:
//  ctxLogger := myLogger.WithCtx(ctx).With("service", "test-service")
//  ctxLogger.Info("hello", "world")
//  // Result is:
//  // ts="2016-07-18T17:15:43Z", level="INFO", traceId="testTrace123", spanId="testSpan123", service="test-service", hello="world"
func (l *Logger) With(keyvals ...interface{}) *Logger {
	return &Logger{
		zaplevel: l.zaplevel,
		logger:   l.logger.With(keyvals...),
	}
}

// SetInfo sets the level of the logger to the INFO level
func (l *Logger) SetInfo() {
	l.SetLevel(INFO)
}

// SetDebug sets the level of the logger to the DEBUG level
func (l *Logger) SetDebug() {
	l.SetLevel(DEBUG)
}

// SetWarning sets the level of the logger to the WARNING level
func (l *Logger) SetWarning() {
	l.SetLevel(WARNING)
}

// SetError sets the level of the logger to the ERROR level
func (l *Logger) SetError() {
	l.SetLevel(ERROR)
}

// SetLevel sets the level of the logger to the specified level
func (l *Logger) SetLevel(lv Level) {
	// flush the buffer
	l.logger.Sync()
	zapLevel := getZapLevel(lv)
	l.zaplevel.SetLevel(zapLevel)
}

// init is a "special" function executed by the Go runtime, and is used to
// initialize the package
func init() {
	err := registerLogfmtEncoder()

	if err != nil {
		panic(err)
	}

	globalLogger = NewLogger(abws.NewAsyncBufferWriteSyncer(os.Stdout, maxBufferSize, maxDelay), INFO)
}

// Info logs a set of keys/values at the INFO level (if enabled)
func Info(keyvals ...interface{}) {
	globalLogger.Info(keyvals...)
}

// Fatal logs an error at the ERROR level, then exits the process with an exit code of 1.
func Fatal(keyvals ...interface{}) {
	globalLogger.Fatal(keyvals...)
}

// Debug logs a set of keys/values at the DEBUG level (if enabled)
func Debug(keyvals ...interface{}) {
	globalLogger.Debug(keyvals...)
}

// Warn logs a set of keys/values at the WARNING level (if enabled)
func Warn(keyvals ...interface{}) {
	globalLogger.Warn(keyvals...)
}

// Error logs a set of keys/values at the ERROR level (if enabled)
func Error(keyvals ...interface{}) {
	globalLogger.Error(keyvals...)
}

// Sync flushes any buffered log entries.  It'd be a good idea to call this on the exit signal event
func Sync() error {
	return globalLogger.Sync()
}

// WithCtx extracts the traceID, spanID, and caller info from the context and adds them to the logging context, and
// returns a copy of the original logger instance with the new context.  GlobalLogger is not changed.
// For example,
//  var ctx = Context.Background()
//
//  // setup tracing information.  (in a middleware)
//	ctx = context.WithValue(ctx, contextdefs.TraceID, "testTrace123")
//  ctx = context.WithValue(ctx, contextdefs.SpanID, "testSpan123")
//  ctx = context.WithValue(ctx, contextdefs.HBOCaller, "pickup")
//
//  llog.WithCtx(ctx).Info("hello", "world", "username", "dcarney")
//
//  Result is:  ts="2017-07-18T17:15:42Z", level="INFO", traceId="testTrace123", spanId="testSpan123", caller="pickup", hello="world", username="dcarney"
func WithCtx(ctx context.Context) (nl *Logger) {
	return globalLogger.WithCtx(ctx)
}

// With allows static values to be added to the llog, so that every log contains these
// values. The original instance remains unchanged.
// For example,
//  moduleLogger := llog.With("service", "test-service")
//  moduleLogger.Info("hello", "world")
//
//  // Result is:
//  // ts="2016-07-18T17:15:42Z", level="INFO",  service="test-service", hello="world"
// This can also be chained for example:
//  ctxLogger := myLogger.WithCtx(ctx).With("service", "test-service")
//  ctxLogger.Info("hello", "world")
//  // Result is:
//  // ts="2016-07-18T17:15:43Z", level="INFO", traceId="testTrace123", spanId="testSpan123", service="test-service", hello="world"
func With(keyvals ...interface{}) (nl *Logger) {
	return globalLogger.With(keyvals...)
}

// SetWriter will create a new globalLogger with specified io.Writer.
func SetWriter(out io.Writer) {
	// Flush buffer
	globalLogger.logger.Sync()
	globalLogger = NewLogger(out, getLlogLevel(globalLogger.zaplevel.Level()))
}

// SetLevel changes the configured logging level of the package
func SetLevel(l Level) {
	// Flush buffer
	globalLogger.SetLevel(l)
}

// getZapLevel converts a llog level to a zapcore.Level
func getZapLevel(l Level) zapcore.Level {
	switch l {
	case DEBUG:
		return zap.DebugLevel
	case INFO:
		return zap.InfoLevel
	case WARNING:
		return zap.WarnLevel
	case ERROR:
		return zap.ErrorLevel
	default:
		return zap.DebugLevel
	}
}

// getZapLevel converts a zapcore.Level to a llog level
func getLlogLevel(l zapcore.Level) Level {
	switch l {
	case zapcore.DebugLevel:
		return DEBUG
	case zapcore.InfoLevel:
		return INFO
	case zapcore.WarnLevel:
		return WARNING
	case zapcore.ErrorLevel:
		return ERROR
	default:
		// The rest of the zapcore.Level are above ERROR (ie.. panic, fatal..) So we translate these to ERROR
		return ERROR
	}
}
