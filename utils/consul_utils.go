package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoowii/jsonrpc_proxygo/config"
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
	body := resp.Body
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("consul KV get %s error %s", url, string(bodyBytes))
		return
	}

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

func ConsulRegisterService(consulConfig *config.ConsulConfig, wholeConfig *config.ServerConfig) (err error) {
	consulUrl := consulConfig.Endpoint
	if strings.Index(consulUrl, "consul://") == 0 {
		consulUrl, err = ConsulUrlToHttpUrl(consulUrl)
		if err != nil {
			return
		}
	}
	consulUrl = strings.TrimSuffix(consulUrl, "/")
	registerUrl := fmt.Sprintf("%s/v1/agent/service/register", consulUrl)
	client := &http.Client{}
	type registerServicePayloadType struct {
		ID                string                 `json:"ID"`
		Name              string                 `json:"Name"`
		Tags              []string               `json:"Tags"`
		Address           *string                `json:"Address"`
		Port              *int                   `json:"Port"`
		Meta              map[string]string      `json:"Meta,omitempty"`
		EnableTagOverride bool                   `json:"EnableTagOverride,omitempty"`
		Check             map[string]interface{} `json:"Check,omitempty"`
		Weights           map[string]int         `json:"Weights,omitempty"`
	}
	checkConf := make(map[string]interface{})
	checkId := fmt.Sprintf("service:%s:health_checker", consulConfig.Id)
	checkConf["CheckID"] = checkId
	consulConfig.HealthCheckId = checkId
	checkConf["TTL"] = "3m" // TODO: 从配置中加载心跳检测的TTL
	payload := &registerServicePayloadType{
		ID:                StringOrElse(consulConfig.Id, "jsonrpc_proxygo_1"),
		Name:              StringOrElse(consulConfig.Name, "jsonrpc_proxygo"),
		Tags:              consulConfig.Tags,
		Address:           wholeConfig.GetEndpointHost(),
		Port:              wholeConfig.GetEndpointPort(),
		Meta:              nil,
		EnableTagOverride: true,
		Check:             checkConf,
		Weights:           nil,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	log.Debugf("start register consul service to url %s", registerUrl)
	req, err := http.NewRequest("PUT", registerUrl, bytes.NewReader(payloadBytes))
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	body := resp.Body
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("register consul service error %s", string(bodyBytes))
		return
	}

	log.Infof("register consul service response %s", string(bodyBytes))
	return
}

func ConsulSubmitHealthChecker(consulConfig *config.ConsulConfig) (err error) {
	consulUrl := consulConfig.Endpoint
	if strings.Index(consulUrl, "consul://") == 0 {
		consulUrl, err = ConsulUrlToHttpUrl(consulUrl)
		if err != nil {
			return
		}
	}
	consulUrl = strings.TrimSuffix(consulUrl, "/")
	checkId := consulConfig.HealthCheckId
	checkUrl := fmt.Sprintf("%s/v1/agent/check/pass/%s", consulUrl, checkId)
	client := &http.Client{}
	req, err := http.NewRequest("PUT", checkUrl, bytes.NewReader(nil))
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	body := resp.Body
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("consul service health check error %s", string(bodyBytes))
		return
	}
	log.Infof("consul service health check response %s", string(bodyBytes))
	return
}
