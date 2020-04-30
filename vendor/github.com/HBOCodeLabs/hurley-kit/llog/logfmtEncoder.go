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
	"encoding/base64"
	"encoding/json"
	"math"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var bufferpool = buffer.NewPool()

// For JSON-escaping; see logfmtEncoder.safeAddString below.
const _hex = "0123456789abcdef"

// In addition to pooling the byte buffers, we need to pool the encoder itself to
// avoid allocating in `EncodeEntry`.
var _logfmtPool = sync.Pool{New: func() interface{} {
	return &logfmtEncoder{}
}}

func getLogfmtEncoder() *logfmtEncoder {
	return _logfmtPool.Get().(*logfmtEncoder)
}

func putLogfmtEncoder(enc *logfmtEncoder) {
	enc.EncoderConfig = nil
	enc.buf = nil
	enc.openNamespaces = 0
	_logfmtPool.Put(enc)
}

type logfmtEncoder struct {
	*zapcore.EncoderConfig
	buf            *buffer.Buffer
	openNamespaces int
}

// NewLogfmtEncoder creates a fast logfmt Encoder. See (https://brandur.org/logfmt)
// The encoder is based on zap's JsonEncoder.
// https://github.com/uber-go/zap/blob/master/zapcore/json_encoder.go
// Note that the encoder doesn't deduplicate keys, so it's possible to produce
// a message like
//   foo=bar foo=baz
func NewLogfmtEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return newLogfmtEncoder(cfg)
}

func newLogfmtEncoder(cfg zapcore.EncoderConfig) *logfmtEncoder {
	return &logfmtEncoder{
		EncoderConfig: &cfg,
		buf:           bufferpool.Get(),
	}
}

// AddArray implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

// AddObject implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

// AddBinary implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

// AddByteString implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

// AddBool implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

// AddComplex128 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

// AddDuration implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

// AddFloat64 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

// AddInt64 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

// AddReflected implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddReflected(key string, obj interface{}) error {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(marshaled)
	return err
}

// OpenNamespace implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) OpenNamespace(key string) {
	enc.addKey(key)
	enc.buf.AppendByte('{')
	enc.openNamespaces++
}

// AddString implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
}

// AddTime implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
}

// AddUint64 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
}

// AddUint64 implements zapcore's ArrayEncoder
func (enc *logfmtEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.addElementSeparator()
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	return err
}

// AppendObject implements zapcore's ArrayEncoder
func (enc *logfmtEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	enc.addElementSeparator()
	enc.buf.AppendByte('{')
	err := obj.MarshalLogObject(enc)
	enc.buf.AppendByte('}')
	return err
}

// AppendBool implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendBool(val bool) {
	enc.addElementSeparator()
	enc.buf.AppendBool(val)
}

// AppendByteString implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendByteString(val []byte) {
	enc.addElementSeparator()
	enc.buf.AppendByte('"')
	enc.safeAddByteString(val)
	enc.buf.AppendByte('"')
}

// AppendComplex128 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendComplex128(val complex128) {
	enc.addElementSeparator()
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendByte('"')
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, 64)
	enc.buf.AppendByte('+')
	enc.buf.AppendFloat(i, 64)
	enc.buf.AppendByte('i')
	enc.buf.AppendByte('"')
}

// AppendDuration implements zapcore's ArrayEncoder
func (enc *logfmtEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	enc.EncodeDuration(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendInt64(int64(val))
	}
}

// AppendInt64 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendInt64(val int64) {
	enc.addElementSeparator()
	enc.buf.AppendInt(val)
}

// AppendReflected implements zapcore's ArrayEncoder
func (enc *logfmtEncoder) AppendReflected(val interface{}) error {
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}
	enc.addElementSeparator()
	_, err = enc.buf.Write(marshaled)
	return err
}

// AppendString implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendString(val string) {
	enc.addElementSeparator()
	enc.buf.AppendByte('"')
	enc.safeAddString(val)
	enc.buf.AppendByte('"')
}

// AppendTime implements zapcore's ArrayEncoder
func (enc *logfmtEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	enc.EncodeTime(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

// AppendUint64 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendUint64(val uint64) {
	enc.addElementSeparator()
	enc.buf.AppendUint(val)
}

// AddComplex64 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }

// AddFloat32 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddFloat32(k string, v float32) { enc.AddFloat64(k, float64(v)) }

// AddInt implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddInt(k string, v int) { enc.AddInt64(k, int64(v)) }

// AddInt32 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddInt32(k string, v int32) { enc.AddInt64(k, int64(v)) }

// AddInt16 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddInt16(k string, v int16) { enc.AddInt64(k, int64(v)) }

// AddInt8 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddInt8(k string, v int8) { enc.AddInt64(k, int64(v)) }

// AddUint implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddUint(k string, v uint) { enc.AddUint64(k, uint64(v)) }

// AddUint32 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddUint32(k string, v uint32) { enc.AddUint64(k, uint64(v)) }

// AddUint16 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddUint16(k string, v uint16) { enc.AddUint64(k, uint64(v)) }

// AddUint8 implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddUint8(k string, v uint8) { enc.AddUint64(k, uint64(v)) }

// AddUintptr implements zapcore's ObjectEncoder
func (enc *logfmtEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }

// AppendComplex64 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendComplex64(v complex64) { enc.AppendComplex128(complex128(v)) }

// AppendFloat64 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendFloat64(v float64) { enc.appendFloat(v, 64) }

// AppendFloat32 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendFloat32(v float32) { enc.appendFloat(float64(v), 32) }

// AppendInt implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendInt(v int) { enc.AppendInt64(int64(v)) }

// AppendInt32 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendInt32(v int32) { enc.AppendInt64(int64(v)) }

// AppendInt16 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendInt16(v int16) { enc.AppendInt64(int64(v)) }

// AppendInt8 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendInt8(v int8) { enc.AppendInt64(int64(v)) }

// AppendUint implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendUint(v uint) { enc.AppendUint64(uint64(v)) }

// AppendUint32 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendUint32(v uint32) { enc.AppendUint64(uint64(v)) }

// AppendUint16 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendUint16(v uint16) { enc.AppendUint64(uint64(v)) }

// AppendUint8 implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendUint8(v uint8) { enc.AppendUint64(uint64(v)) }

// AppendUintptr implements zapcore's PrimitiveArrayEncoder
func (enc *logfmtEncoder) AppendUintptr(v uintptr) { enc.AppendUint64(uint64(v)) }

// Clone implements zapcore's Encoder, copies the encoder, ensuring that adding fields to the copy doesn't affect the original.
func (enc *logfmtEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *logfmtEncoder) clone() *logfmtEncoder {
	clone := getLogfmtEncoder()
	clone.EncoderConfig = enc.EncoderConfig
	clone.openNamespaces = enc.openNamespaces
	clone.buf = bufferpool.Get()
	return clone
}

// EncodeEntry implements zapcore's Encoder.  It encodes an entry and fields, along with any accumulated context, into a byte buffer
// returns it.  The main difference from zap's jsonEncoder is that timestamp is at the beginning
// https://answers.splunk.com/answers/1951/what-is-the-best-custom-log-event-format-for-splunk-to-eat.html
// ts="2016-05-18 21:13:02.956 UTC", level="INFO", apiName="post.wombats", url="/wombats", method="POST", httpStatus=201,
// elapsed=5, event="serverSend", traceId="cda704cfa1c08d35", spanId="cda704cfa1c08d35",
// service=perf line=serviceTrace.js:204 src=/home/dcarney/src/Hurley-Perf/node_modules/@hbo/hurley-common/lib/util/serviceTrace.js:204:serverSend
func (enc *logfmtEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()

	// Putting timestamp at the beginning because it's splunk friendly
	if final.TimeKey != "" {
		final.AddTime(final.TimeKey, ent.Time)
	}

	if final.LevelKey != "" {
		final.addElementSeparator()
		final.addKey(final.LevelKey)
		cur := final.buf.Len()
		final.EncodeLevel(ent.Level, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeLevel was a no-op. Fall back to strings to keep
			// output JSON valid.
			final.AppendString(ent.Level.String())
		}
	}

	if ent.LoggerName != "" && final.NameKey != "" {
		final.AppendString(ent.LoggerName)
	}

	if ent.Caller.Defined && final.CallerKey != "" {
		final.addKey(final.CallerKey)
		cur := final.buf.Len()
		final.EncodeCaller(ent.Caller, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeCaller was a no-op. Fall back to strings to
			// keep output JSON valid.
			final.AppendString(ent.Caller.String())
		}
	}
	if final.MessageKey != "" {
		final.addKey(enc.MessageKey)
		final.AppendString(ent.Message)
	}
	if enc.buf.Len() > 0 {
		final.addElementSeparator()
		final.buf.Write(enc.buf.Bytes())
	}
	addFields(final, fields)
	final.closeOpenNamespaces()
	if ent.Stack != "" && final.StacktraceKey != "" {
		final.AddString(final.StacktraceKey, ent.Stack)
	}
	if final.LineEnding != "" {
		final.buf.AppendString(final.LineEnding)
	} else {
		final.buf.AppendString(zapcore.DefaultLineEnding)
	}

	ret := final.buf
	putLogfmtEncoder(final)
	return ret, nil
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

func (enc *logfmtEncoder) truncate() {
	enc.buf.Reset()
}

func (enc *logfmtEncoder) closeOpenNamespaces() {
	for i := 0; i < enc.openNamespaces; i++ {
		enc.buf.AppendByte('}')
	}
}

// addKey adds the key to the buffer.  The difference from jsonEncoder is it doesn't wrap the key in double quotes, and
// it uses = instead of :
func (enc *logfmtEncoder) addKey(key string) {
	enc.addElementSeparator()
	enc.safeAddString(key)
	enc.buf.AppendByte('=')
}

// addElementSeparator adds a space after certain characters.  The difference from jsonEncoder is it includes = to be
// the ignored list, and that the default separator is ", "
func (enc *logfmtEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	switch enc.buf.Bytes()[last] {
	case '{', '[', ':', ',', ' ', '=':
		return
	default:
		enc.buf.AppendString(", ")
	}
}

func (enc *logfmtEncoder) appendFloat(val float64, bitSize int) {
	enc.addElementSeparator()
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *logfmtEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

// safeAddByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (enc *logfmtEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.Write(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (enc *logfmtEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte(b)
	case '\n':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('n')
	case '\r':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('r')
	case '\t':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (enc *logfmtEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}
