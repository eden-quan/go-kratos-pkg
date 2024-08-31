package middlewarepkg

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/transport"

	errorpkg "github.com/eden-quan/go-kratos-pkg/error"
)

// Server is an server logging middleware.
func Server(logger log.Logger) middleware.Middleware {
	logHandler := log.NewHelper(logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int32
				bizCode   int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}

			reply, err = handler(ctx, req)

			if info, e := errorpkg.NewErrorMetaInfo(err); e == nil {
				leaf := info.Leaf()
				code = leaf.Code
				reason = leaf.Reason
				bizCode = leaf.BizCode
			}

			level, stack := extractError(err)
			// _ = log.WithContext(ctx, logger).Log(level,
			logHandler.WithContext(ctx).Log(level,
				"kind", "server",
				"component", kind,
				"operation", operation,
				"args", extractArgs(req),
				"biz_code", bizCode,
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// ClientLogging is a client logging middleware.
func ClientLogging(logger log.Logger) middleware.Middleware {
	logHandler := log.NewHelper(logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int32
				bizCode   int32
				reason    string
				kind      string
				operation string
				stack     string
				level     log.Level
			)
			startTime := time.Now()
			if info, ok := transport.FromClientContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			reply, err = handler(ctx, req)

			if info, e := errorpkg.NewErrorMetaInfo(err); e == nil {
				leaf := info.Leaf()
				code = leaf.Code
				reason = leaf.Reason
				bizCode = leaf.BizCode
				stack = info.Leaf().Reason
				level = leaf.LogLevel()
			} else {
				level, stack = extractError(err)
			}

			// _ = log.WithContext(ctx, logger).Log(level,
			logHandler.WithContext(ctx).Log(level,
				"kind", "client",
				"component", kind,
				"operation", operation,
				"args", extractArgs(req),
				"code", code,
				"biz_code", bizCode,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if redacter, ok := req.(logging.Redacter); ok {
		return redacter.Redact()
	}
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, fmt.Sprintf("%+v", err)
	}
	return log.LevelInfo, ""
}
