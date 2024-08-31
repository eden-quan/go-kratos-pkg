package apppkg

import (
	stdjson "encoding/json"
	errorpkg "github.com/eden-quan/go-kratos-pkg/error"
	headerpkg "github.com/eden-quan/go-kratos-pkg/header"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/tidwall/sjson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	stdhttp "net/http"
	"strings"
)

const (
	OK = 0

	baseContentType = "application"
)

var _ = http.DefaultResponseEncoder

// ResponseEncoder http.DefaultResponseEncoder
func ResponseEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, v interface{}) error {
	// 在websocket时日志干扰：http: superfluous response.WriteHeader call from xxx(file:line)
	// 在websocket时日志干扰：http: response.Write on hijacked connection from
	// is websocket
	if headerpkg.GetIsWebsocket(r.Header) {
		return nil
	}

	// nil
	if v == nil {
		//respData := &Response{
		//	Code:      OK,
		//	RequestId: headerpkg.GetRequestID(r.Header),
		//	//Data:      v,
		//}
		//respData.Code = stdhttp.StatusInternalServerError
		//respData.Reason = errorv1.ERROR_NO_CONTENT.String()
		//respData.Metadata = map[string]string{"data": "null"}
		return nil
	}

	// 响应
	if rd, ok := v.(http.Redirector); ok {
		url, code := rd.Redirect()
		stdhttp.Redirect(w, r, url, code)
		return nil
	}

	// 响应结果
	respData := &Response{
		Code: OK,
		//RequestId: headerpkg.GetRequestID(r.Header),
		//Data:      v,
	}
	var resultMessage proto.Message
	if vMessage, ok := v.(proto.Message); ok {
		// message
		resultMessage = vMessage
	} else {
		// unknown
		vBytes, _ := stdjson.Marshal(v)
		resultMessage = &ResponseData{
			Data: string(vBytes),
		}
	}
	anyData, err := anypb.New(resultMessage)
	if err != nil {
		respData.Code = stdhttp.StatusInternalServerError
		respData.Reason = "INTERNAL_SERVER"
		respData.Metadata = map[string]string{"error": err.Error()}
	} else {
		respData.Data = anyData
	}

	// return
	codec, _ := http.CodecForRequest(r, "Accept")
	SetResponseContentType(w, codec)
	w.WriteHeader(stdhttp.StatusOK)

	// return
	dataBytes, err := codec.Marshal(respData)
	if err != nil {
		return err
	}
	// 删除 data.@type
	newDataByte, err := DeleteDataTypeURL(dataBytes)
	if err != nil {
		_, err = w.Write(dataBytes)
		if err != nil {
			return err
		}
	} else {
		_, err = w.Write(newDataByte)
		if err != nil {
			return err
		}
	}
	return nil

	// 参考
	//return http.DefaultResponseEncoder(w, r, respData)
}

var _ = http.DefaultErrorEncoder

// ErrorEncoder http.DefaultErrorEncoder
func ErrorEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
	// 在websocket时日志干扰：http: superfluous response.WriteHeader call from xxx(file:line)
	// 在websocket时日志干扰：http: response.Write on hijacked connection from
	// is websocket
	if headerpkg.GetIsWebsocket(r.Header) {
		return
	}

	// 响应错误
	se := errorpkg.FromError(err)
	data := &Response{
		Code:     se.Code,
		Reason:   se.Reason,
		Message:  se.Message,
		Metadata: se.Metadata,
		//RequestId: headerpkg.GetRequestID(r.Header),
	}
	if !IsDebugMode() {
		data.Metadata = nil
	}

	codec, _ := http.CodecForRequest(r, "Accept")
	SetResponseContentType(w, codec)

	// // return
	//body, err := codec.Marshal(se)
	body, err := codec.Marshal(data)
	if err != nil {
		w.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(stdhttp.StatusOK)
	//w.WriteHeader(int(se.Code))
	_, _ = w.Write(body)

	// 参考
	//_ = http.DefaultResponseEncoder(w, r, data)
	//http.DefaultErrorEncoder(w, r, err)
	return
}

// ContentType returns the content-type with base prefix.
func ContentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}

// SetResponseContentType ...
func SetResponseContentType(w stdhttp.ResponseWriter, codec encoding.Codec) {
	switch codec.Name() {
	case json.Name:
		w.Header().Set("Content-Type", headerpkg.ContentTypeJSONUtf8)
	default:
		w.Header().Set("Content-Type", ContentType(codec.Name()))
	}
}

// DeleteDataTypeURL ...
func DeleteDataTypeURL(buf []byte) ([]byte, error) {
	p := "data.\\@type"
	res, err := sjson.DeleteBytes(buf, p)
	if err != nil {
		return nil, err
	}
	return res, nil
	//if r := gjson.GetBytes(buf, p); r.Exists() {
	//	res, err := sjson.DeleteBytes(buf, p)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return res, nil
	//}
	//return buf, nil
}
