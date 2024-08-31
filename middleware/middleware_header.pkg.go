package middlewarepkg

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/trace"

	headerpkg "github.com/eden/go-kratos-pkg/header"
)

// RequestAndResponseHeader 请求头 and 响应头
func RequestAndResponseHeader() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var traceID string
			span := trace.SpanFromContext(ctx)
			if span.SpanContext().IsValid() {
				traceID = span.SpanContext().TraceID().String()
			}

			tr, ok := transport.FromServerContext(ctx)
			if ok {
				//tr.ReplyHeader().Set(headerpkg.TraceID, span.SpanContext().TraceID().String())
				if tr.ReplyHeader().Get(headerpkg.TraceID) == "" {
					tr.ReplyHeader().Set(headerpkg.TraceID, traceID)
				}
				if tr.RequestHeader().Get(headerpkg.RequestID) == "" {
					tr.RequestHeader().Set(headerpkg.RequestID, traceID)
				}
			}
			return handler(ctx, req)
		}
	}
}
