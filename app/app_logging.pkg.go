package apppkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport/http"

	contextpkg "github.com/eden/go-kratos-pkg/context"
	errorpkg "github.com/eden/go-kratos-pkg/error"
	headerpkg "github.com/eden/go-kratos-pkg/header"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

var (
	// _maxRequestArgs 设置最大请求参数
	_maxRequestArgs uint = 1024 * 1024
)

// SetMaxRequestArgSize 设置最大请求参数
func SetMaxRequestArgSize(size uint) {
	_maxRequestArgs = size
}

// RequestMessage 请求信息
type RequestMessage struct {
	Kind      string
	Component string
	Method    string
	Operation string

	ExecTime time.Duration
	ClientIP string
}

// GetRequestInfoSlice 获取本次调用信息信息
func (s *RequestMessage) GetRequestInfoSlice() []interface{} {
	return []interface{}{
		"request.kind", s.Kind,
		"request.component", s.Component,
		"request.ip", s.ClientIP,
		"request.latency", s.ExecTime.Milliseconds(),
		"request.method", s.Method,
		"request.operation", s.Operation,
	}
}

// ErrMessage 响应信息
type ErrMessage struct {
	Code    int32
	BizCode int32
	Reason  string
	Msg     string
	Stack   string

	RequestArgs string
}

// GetErrorDetailSlice 获取错误信息，当无错误时，error.code 为 0
func (s *ErrMessage) GetErrorDetailSlice() []interface{} {
	return []interface{}{
		"error.code", s.Code,
		"error.biz_code", s.BizCode,
		"error.reason", s.Reason,
		"error.detail", s.Msg,
		"error.stack", s.Stack,
		"error.args", s.RequestArgs,
	}
}

// options ...
type options struct {
	withSkip      bool
	withSkipDepth int
}

// Option ...
type Option func(options *options)

// WithDefaultSkip ...
func WithDefaultSkip() Option {
	return func(o *options) {
		o.withSkip = true
		o.withSkipDepth = 1
	}
}

// WithCallerSkip ...
func WithCallerSkip(skip int) Option {
	return func(o *options) {
		o.withSkip = true
		o.withSkipDepth = skip
	}
}

// ServerLog 中间件日志
// 参考 logging.Server(logger)
func ServerLog(logger log.Logger, opts ...Option) middleware.Middleware {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				isWebsocket    = false
				loggingLevel   = log.LevelInfo
				requestMessage = &RequestMessage{
					Kind: "server",
				}
				errMessage = &ErrMessage{
					Code: 0,
				}
			)

			// 信息
			tr, ok := transport.FromServerContext(ctx)
			if ok {
				requestMessage.Component = tr.Kind().String()
				requestMessage.Operation = tr.Operation()
			}

			// 时间
			startTime := time.Now()

			// 执行结果
			reply, err = handler(ctx, req)

			// 执行时间
			requestMessage.ExecTime = time.Since(startTime)

			// request
			if httpTr, isHTTP := tr.(http.Transporter); isHTTP {
				requestMessage.Method = httpTr.Request().Method
				requestMessage.Operation = httpTr.Request().URL.String()
				if headerpkg.GetIsWebsocket(httpTr.Request().Header) {
					isWebsocket = true
					requestMessage.Method = "WS"
				}
				requestMessage.ClientIP = contextpkg.ClientIPFromHTTP(ctx, httpTr.Request())
			} else {
				requestMessage.Method = "GRPC"
				requestMessage.ClientIP = contextpkg.ClientIPFromGRPC(ctx)
			}

			// 本次请求调用信息
			kv := requestMessage.GetRequestInfoSlice()

			// websocket 不输出错误
			if isWebsocket {
				_ = log.WithContext(ctx, logger).Log(loggingLevel, kv...)
				return
			}

			if !errorpkg.IsEmptyError(err) {
				loggingLevel = log.LevelError
				// 错误信息
				if info, e := errorpkg.NewErrorMetaInfo(err); e == nil {
					leaf := info.Leaf()
					errMessage.Code = leaf.Code
					errMessage.Reason = leaf.Reason
					errMessage.BizCode = leaf.BizCode
					errMessage.Stack = leaf.Stack
					errMessage.Msg = info.Error()
					loggingLevel = leaf.LogLevel()

					// keep it for other middleware
					//err = leaf.Error
				} else {
					// 未处理的错误
					var callers []string
					if opt.withSkip && opt.withSkipDepth > 0 {
						callers = errorpkg.CallerWithSkip(err, opt.withSkipDepth)
					} else {
						callers = errorpkg.Stack(err)
					}
					if len(callers) > 0 {
						errMessage.Stack = strings.Join(callers, "\n\t")
					}

					errMessage.Msg = err.Error()
					errMessage.Msg = "WARNING [UNDEFINED ERROR] " + errMessage.Msg
				}

				// 请求参数
				errMessage.RequestArgs = extractArgs(req)
				if len(errMessage.RequestArgs) > int(_maxRequestArgs) {
					errMessage.RequestArgs = errMessage.RequestArgs[:_maxRequestArgs]
				}
			}
			// 补充错误日志信息
			kv = append(kv, errMessage.GetErrorDetailSlice()...)

			// 输出日志
			_ = log.WithContext(ctx, logger).Log(loggingLevel, kv...)
			return
		}

	}
}

// ClientLog is an client logging middleware.
// logging.Client(logger)
func ClientLog(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int32
				reason    string
				bizCode   int32
				kind      string
				operation string
			)
			startTime := time.Now()
			if info, ok := transport.FromClientContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			reply, err = handler(ctx, req)
			if info, err := errorpkg.NewErrorMetaInfo(err); err == nil {
				leaf := info.Leaf()
				code = leaf.Code
				reason = leaf.Reason
				bizCode = leaf.BizCode
			}

			level, stack := extractError(err)
			_ = log.WithContext(ctx, logger).Log(level,
				"kind", "client",
				"component", kind,
				"operation", operation,
				"args", extractArgs(req),
				"code", code,
				"reason", reason,
				"biz_code", bizCode,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
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
