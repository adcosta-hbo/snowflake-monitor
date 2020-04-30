// Package defs contains definitions for common span and tag names used in
// tracing.
//
// These definitions extend the various tag and span name
// definitions that are already a part of the opentracing/ext package:
// https://github.com/opentracing/opentracing-go/blob/master/ext/tags.go
//
// By using consistent span and tag names, our overall tracing system will
// be easier to navigate, search, and reason about.
//
// For more details about naming conventions, see:
// http://opentracing.io/documentation/pages/api/data-conventions.html
package defs

import "net/http"

const (

	// HTTPServer is the string used to name the spans created for each
	// incoming HTTP request.
	HTTPServer string = "http_server"

	// HTTPClient is the string used to name the spans created for each
	// outgoing HTTP request.
	HTTPClient string = "http_client"

	// SQSWrite is used to name/tag a span for writing to an SQS queue
	SQSWrite string = "sqs.write"

	// SQSRead is used to name/tag a span for reading from an SQS queue
	SQSRead string = "sqs.read"

	// SQSQueueName is used as a tag key, when providing the name of an SQS queue
	SQSQueueName string = "sqs.queue_name"

	// RedisScriptEval is used to name/tag a span for excuting a redis Lua script
	RedisScriptEval string = "redis.eval"

	// RedisCmd is used to name/tag a span for excuting a command against a redis server
	RedisCmd string = "redis.cmd"

	// DynamoRead is used to name/tag a span for executing a read operation against an AWS Dynamo database
	DynamoRead string = "dynamo.read"

	// DynamoWrite is used to name/tag a span for executing a write operation against an AWS Dynamo database
	DynamoWrite string = "dynamo.write"

	// DynamoTableName is used as a tag key, when providing the name of the Dynamo table
	DynamoTableName string = "dynamo.table_name"
)

var (
	// TraceIDHeaderName is the name of the HTTP header that is used to store/propagate
	// the trace ID value.
	TraceIDHeaderName = http.CanonicalHeaderKey("X-B3-TraceID")

	// SpanIDHeaderName is the name of the HTTP header that is used to store/propagate
	// the span ID value.
	SpanIDHeaderName = http.CanonicalHeaderKey("X-B3-SpanID")

	// ParentSpanIDHeaderName is the name of the HTTP header that is used to store/propagate
	// the parent span ID value.
	ParentSpanIDHeaderName = http.CanonicalHeaderKey("X-B3-ParentSpanID")

	// UberOpentracingHeaderName is the name of the HTTP header for opentracing information
	UberOpentracingHeaderName = http.CanonicalHeaderKey("uber-trace-id")
)
