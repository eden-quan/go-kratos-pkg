package apppkg

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc/codes"
	stdhttp "net/http"
)

var (
	_ = http.DefaultRequestDecoder
	_ = http.DefaultErrorEncoder
	_ = http.DefaultResponseEncoder
	_ = http.DefaultResponseDecoder
)

// HTTPResponseInterface .
type HTTPResponseInterface interface {
	GetCode() int32
	GetReason() string
	GetMessage() string
	GetMetadata() map[string]string
}

// HTTPResponse 响应
// 关联更新 apppkg.Response
type HTTPResponse struct {
	Code     int32             `json:"code"`
	Reason   string            `json:"reason"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`

	Data interface{} `json:"data"`
}

func (x *HTTPResponse) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *HTTPResponse) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *HTTPResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *HTTPResponse) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

// IsSuccessCode 成功的响应码
func IsSuccessCode(code int32) bool {
	if code == OK {
		return true
	}
	return IsSuccessHTTPCode(int(code))
}

// IsSuccessHTTPCode 成功的HTTP响应吗
func IsSuccessHTTPCode(code int) bool {
	if code >= stdhttp.StatusOK && code < stdhttp.StatusMultipleChoices {
		return true
	}
	return false
}

// IsSuccessGRPCCode 成功的GRPC响应吗
func IsSuccessGRPCCode(code uint32) bool {
	return codes.Code(code) == codes.OK
}

// ToResponseError 转换为错误
func ToResponseError(response HTTPResponseInterface) *errors.Error {
	return &errors.Error{
		Status: errors.Status{
			Code:     response.GetCode(),
			Reason:   response.GetReason(),
			Message:  response.GetMessage(),
			Metadata: response.GetMetadata(),
		},
	}
}
