package myMock

import (
	"errors"
	"net/http"
	"wcj-go-text/golang/myText"

	"github.com/spf13/viper"
)

type Mock struct {
	Name    string
	Resp    string
	ReqType string
}

func (m Mock) BuildMockResp() string {
	rp := myText.Replacer{
		Template: m.Resp,
	}
	rp1 := rp.RepalceText()
	//2次替换
	rp2 := myText.Replacer{
		Template: rp1,
	}
	return rp2.RepalceText()
}

// 请求头匹配
func mappingHeader(headMap map[string]interface{}, r *http.Request) bool {
	mapping := true
	for hk, hv := range headMap {
		if r.Header.Get(hk) != hv {
			return false
		}
	}
	return mapping
}

func BuildMock(path string, r *http.Request, mockNode string) (Mock, error) {
	var mock Mock
	config := viper.GetStringMap(mockNode)
	for k, v := range config {
		mockValue := v.(map[string]interface{})
		//url 匹配
		if mockValue["mapping"].(string) != path {
			continue
		}
		//header匹配
		if mockValue["mapping_header"] != nil {
			headMap := mockValue["mapping_header"].(map[string]interface{})
			if !mappingHeader(headMap, r) {
				continue
			}
		}

		mock = Mock{
			Name:    k,
			Resp:    mockValue["resp"].(string),
			ReqType: mockValue["type"].(string),
		}
		return mock, nil
	}
	return mock, errors.New("没有找到匹配的mock")
}
