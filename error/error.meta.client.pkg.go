package errorpkg

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/errors"
)

type ClientMarshalInfo struct {
	BizCode        int               `json:"biz_code"`
	Code           int               `json:"code"`
	DefaultMessage string            `json:"default_message"`
	Message        string            `json:"message"`
	MessageStack   string            `json:"message_stack"`
	MetaData       map[string]string `json:"meta_data"`
}

func NewClientMarshalInfo(raw error) *ClientMarshalInfo {
	c := &ClientMarshalInfo{}

	if raw == nil {
		return c
	}

	err := new(errors.Error)
	if !errors.As(raw, &err) {
		c.Message = raw.Error()
		return c
	}

	m := MetaFromError(err)
	c.BizCode = int(m.BizCode)
	c.Code = int(m.Code)
	c.DefaultMessage = m.DefaultMessage
	c.Message = m.Message
	c.MetaData = m.Metadata

	return c
}

func (c *ClientMarshalInfo) Json() string {
	b, _ := json.Marshal(c)
	return string(b)
}
