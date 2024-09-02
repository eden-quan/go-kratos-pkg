package errorpkg

import (
	"strconv"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type MetaInfo struct {
	*errors.Error
	BizCode        int32
	DefaultMessage string
	Stack          string
}

func MetaFromError(raw error) *MetaInfo {
	if raw == nil {
		return nil
	}

	err := new(errors.Error)
	if !errors.As(raw, &err) {
		// try status again, if success means its from client side
		status := errors.FromError(raw)
		if _, ok := status.Metadata[BizCodeKey]; ok {
			err = status
		} else {
			return nil
		}
	}

	bizCode, _ := strconv.Atoi(err.Metadata[BizCodeKey])

	info := &MetaInfo{
		BizCode:        int32(bizCode),
		DefaultMessage: err.Metadata[DefaultMessageKey],
		Stack:          err.Metadata[StackKey],
		Error:          err,
	}

	if info.Message == "" {
		info.Message = info.DefaultMessage
	}

	return info
}

func (e *MetaInfo) LogLevel() log.Level {
	if e.Code < 500 || e.Code > 600 {
		return log.LevelWarn
	}
	return log.LevelError
}

func (e *MetaInfo) CleanError() *errors.Error {
	err := errors.Clone(e.Error)
	if _, ok := err.Metadata[MetaDataKey]; ok {
		delete(err.Metadata, MetaDataKey)
	}
	if _, ok := err.Metadata[StackKey]; ok {
		delete(err.Metadata, StackKey)
	}

	return err
}

func (e *MetaInfo) String() string {
	return e.Message
}
