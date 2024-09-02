package errorpkg

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
)

const (
	MetaDataKey       = "__MetaKey"
	StackKey          = "__Stack"
	BizCodeKey        = "BizCode"
	DefaultMessageKey = "DefaultMessage"
)

func IsEmptyError(err error) bool {
	if err == nil {
		return true
	}

	error2, ok := interface{}(err).(*errors.Error)
	if ok && error2 == nil {
		return true
	}

	return false
}

type ErrorMetaInfo struct {
	BaseError    error
	WrapperStack []*MetaInfo
}

func NewErrorMetaInfo(err error) (*ErrorMetaInfo, error) {
	if IsEmptyError(err) {
		return nil, fmt.Errorf("err is nil")
	}
	info := &ErrorMetaInfo{
		BaseError:    err,
		WrapperStack: make([]*MetaInfo, 0),
	}

	anaErr := info.analyse()
	return info, anaErr
}

type errorWrap struct {
	err     error
	isError bool
}

func (e *errorWrap) toError() *errors.Error {
	se := new(errors.Error)
	errors.As(e.err, &se)
	return se
}

func (e *errorWrap) toErr() error {
	return e.err
}

// errorChain 清理错误链中的元信息只保留错误，然后构建错误的封装链路
func (e *ErrorMetaInfo) errorChain() []*errorWrap {
	err := e.ClearMeta()
	msg := make([]*errorWrap, 0)

	for err != nil {
		if se := new(errors.Error); errors.As(err, &se) {
			meta := maps.Clone(se.Metadata)
			if _, ok := meta[StackKey]; ok {
				delete(meta, StackKey)
			}
			msg = append(msg, &errorWrap{err: err, isError: true})
			err = se.Unwrap()
		} else {
			msg = append(msg, &errorWrap{err: err, isError: false})
			err = nil
		}
	}

	slices.Reverse(msg)
	return msg
}

func (e *ErrorMetaInfo) addInfo(err *errors.Error) *MetaInfo {
	info := MetaFromError(err)
	e.WrapperStack = append(e.WrapperStack, info)

	return info
}

func (e *ErrorMetaInfo) analyse() error {
	rootSe := new(errors.Error)
	if !errors.As(e.BaseError, &rootSe) {
		// try status again, if success means it't from client side
		status := errors.FromError(e.BaseError)
		if _, ok := status.Metadata[BizCodeKey]; ok {
			_ = e.addInfo(status)
			return nil
		}

		return fmt.Errorf("error doesn't match errors.Error | status.Status")
	}

	// find most deep se
	se := rootSe
	for {
		// if se's cause is meta, try to find next level
		deep := se.Unwrap()
		if deep == nil {
			_ = e.addInfo(se)
		} else if deepSe := new(errors.Error); errors.As(deep, &deepSe) {

			if _, isMeta := deepSe.Metadata[MetaDataKey]; isMeta {
				info := e.addInfo(deepSe)
				se.Metadata[BizCodeKey] = strconv.Itoa(int(info.BizCode))
				se.Metadata[DefaultMessageKey] = info.DefaultMessage
				for k, v := range se.Metadata {
					deepSe.Metadata[k] = v
				}
			}

			deep = deepSe.Unwrap()
			if errors.As(deep, &deepSe) {
				se = deepSe
				continue
			}
		}
		break
	}

	return nil
}

func (e *ErrorMetaInfo) Error() string {
	errorChain := e.errorChain()
	msg := make([]string, 0, len(errorChain))

	for _, m := range errorChain {
		if m.isError {
			err := m.toError()
			msg = append(msg, fmt.Sprintf("error: code %d reason = %s mesage = %s metadata = %v", err.Code, err.Reason, err.Message, err.Metadata))
		} else {
			err := m.toErr()
			msg = append(msg, fmt.Sprintf("error: %s", err.Error()))
		}
	}

	return strings.Join(msg, "\n")
}

func (e *ErrorMetaInfo) ErrorStack() string {
	// err := e.ClearMeta()
	errorChain := e.errorChain()
	msg := make([]string, 0, len(errorChain))

	for _, m := range errorChain {
		if m.isError {
			err := m.toError()
			msg = append(msg, fmt.Sprintf("error: code %d reason = %s mesage = %s", err.Code, err.Reason, err.Message))
		} else {
			err := m.toErr()
			msg = append(msg, fmt.Sprintf("error: %s", err.Error()))
		}
	}

	return strings.Join(msg, "\n")
}

func (e *ErrorMetaInfo) Leaf() *MetaInfo {
	if len(e.WrapperStack) == 0 {
		return nil
	}

	return e.WrapperStack[len(e.WrapperStack)-1]
}

// ClearMeta 接收一个错误，如果他的 cause 是 Meta, 则创建一个新的 error, 用 cause 的 cause 来作为自己的 cause
func (e *ErrorMetaInfo) ClearMeta() error {
	return ClearMeta(e.BaseError)
}

// ToClientError 由 Server 段调用，生成一个包含于服务端相同错误信息的基础错误
func (e *ErrorMetaInfo) ToClientError() error {
	leaf := e.Leaf()
	msg := e.Error()
	// 清理错误堆栈
	meta := maps.Clone(leaf.Metadata)
	for _, k := range []string{StackKey, MetaDataKey} {
		if _, ok := meta[k]; ok {
			delete(meta, k)
		}
	}

	// TODO: 判断是否要合并各层的 MetaData
	return errors.New(int(leaf.Code), leaf.Reason, msg).WithMetadata(meta)
}

// ClearMeta 接收一个错误，如果他的 cause 是 Meta, 则创建一个新的 error, 用 cause 的 cause 来作为自己的 cause
func ClearMeta(rawErr error) error {
	if rawErr == nil {
		return nil
	}

	err := new(errors.Error)
	if !errors.As(rawErr, &err) {
		return rawErr
	}

	inner := err.Unwrap()
	if innerErr := new(errors.Error); errors.As(inner, &innerErr) {
		// replace cause if its meta
		if _, isInnerMeta := innerErr.Metadata[MetaDataKey]; isInnerMeta {
			// reset innerErr to innerErr.cause
			inner = innerErr.Unwrap()
			inner = ClearMeta(inner)
		}
		return err.WithCause(ClearMeta(inner))
	} else {
		return err
	}
}
