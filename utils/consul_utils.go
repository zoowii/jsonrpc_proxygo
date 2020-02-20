package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type ConsulKVPair struct {
	Key         string `json:"Key"`
	Value       string `json:"Value"`
	CreateIndex int    `json:"CreateIndex"`
	ModifyIndex int    `json:"ModifyIndex"`
}

// 把 consul:// 开头的url转换成 http://开头的url
func ConsulUrlToHttpUrl(consulUrl string) (result string, err error) {
	uri, err := url.ParseRequestURI(consulUrl)
	if err != nil {
		return
	}
	uri.Scheme = "http"
	result = uri.String()
	return
}

// 从consul KV url 中获取结果
func ConsulGetKV(url string) (result *ConsulKVPair, err error) {
	if strings.Index(url, "consul://") == 0 {
		url, err = ConsulUrlToHttpUrl(url)
		if err != nil {
			return
		}
	}
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("consul KV get %s error %s", url, resp.Status)
		return
	}
	body := resp.Body
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	var pairs []*ConsulKVPair
	err = json.Unmarshal(bodyBytes, &pairs)
	if err != nil {
		err = fmt.Errorf("consul KV get %s unmarshal json error %s", url, string(bodyBytes))
		return
	}
	if len(pairs) < 1 {
		err = fmt.Errorf("consul KV get %s unmarshal json error %s", url, string(bodyBytes))
		return
	}
	pair := pairs[0]
	base64Value := pair.Value
	valueBytes, err := base64.StdEncoding.DecodeString(base64Value)
	if err != nil {
		return
	}
	value := string(valueBytes)
	pair.Value = value
	result = pair
	return
}
