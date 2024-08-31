package apppkg

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"
)

func getResponseBuffer() []byte {
	data := &Response{Message: "testdata"}
	anyData, err := anypb.New(data)
	if err != nil {
		panic(err)
	}
	res := &Response{Code: 111, Message: "res", Data: anyData}
	buf, err := MarshalJSON(res)
	if err != nil {
		panic(err)
	}
	return buf
}

// go test -v -count=2 ./app -bench=Benchmark_GO_JSON -run=Benchmark_GO_JSON
func Benchmark_GO_JSON(b *testing.B) {
	buf := getResponseBuffer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := make(map[string]interface{})
		err := json.Unmarshal(buf, &res)
		if err != nil {
			b.Fatal("json.Unmarshal(buf, &res)", err)
		}
		data, ok := res["data"].(map[string]interface{})
		if !ok {
			b.Fatal("res[data].(map[string]interface{})")
		}
		if _, ok := data["@type"]; ok {
			delete(data, "@type")
		}
		_, err = json.Marshal(res)
		if err != nil {
			b.Fatal("json.Marshal(res)", err)
		}
	}
}

// go test -v -count=2 ./app -bench=Benchmark_SJSON -run=Benchmark_SJSON
func Benchmark_SJSON(b *testing.B) {
	buf := getResponseBuffer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := "data.\\@type"
		if r := gjson.GetBytes(buf, p); r.Exists() {
			_, err := sjson.DeleteBytes(buf, p)
			if err != nil {
				b.Fatal("json.Marshal(res)", err)
			}
		}
	}
}

// go test -v -count=1 ./app -test.run=Test_DiscardAType
func Test_DiscardAType(t *testing.T) {
	data := &Response{Message: "testdata"}
	anyData, err := anypb.New(data)
	t.Log("anyData.GetTypeUrl: ", anyData.GetTypeUrl())

	res := &Response{Code: 111, Message: "res", Data: anyData}

	buf, err := MarshalJSON(res)
	require.Nil(t, err)
	t.Log("res.String: ", string(buf))

	jRes := gjson.GetBytes(buf, "data.@type")
	t.Log("jRes.String: ", jRes.String())

	p := "data.\\@type"
	if r := gjson.GetBytes(buf, p); r.Exists() {
		buf, err = sjson.DeleteBytes(buf, p)
		require.Nil(t, err)
	}
	t.Log("after delete buf: ", string(buf))

	buf1, err := sjson.DeleteBytes(buf, "fake_key")
	require.Nil(t, err)
	t.Log("after delete buf1: ", string(buf1))
}
