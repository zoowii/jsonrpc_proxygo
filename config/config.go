package config

import (
	"encoding/json"
)

type ServerConfig struct {
	Endpoint string `json:"endpoint"`

	Log struct {
		Level string `json:"level,omitempty"` // DEBUG,INFO,WARN,ERROR, INFO is default
		OutputFile string `json:"output_file,omitempty"`
	} `json:"log,omitempty"`

	Plugins struct {
		// upstream plugin config
		Upstream struct {
			TargetEndpoints []struct{
				Url string `json:"url"`
				Weight int64 `json:"weight"`
			} `json:"upstream_endpoints"`
		} `json:"upstream,omitempty"`

		// cache plugin config
		Caches []struct {
			Name string `json:"name"`
			ParamsForCache []interface{} `json:"paramsForCache"`
			ExpireSeconds int64 `json:"expire_seconds"`
		} `json:"caches,omitempty"`

		BeforeCacheConfigs []struct {
			MethodName string `json:"method"`
			FetchCacheKeyFromParamsCount int `json:"fetch_cache_key_from_params_count"`
		} `json:"before_cache_configs,omitempty"`

		Statistic struct {
			Start bool `json:"start,omitempty"`
		} `json:"statistic,omitempty"`
	} `json:"plugins,omitempty"`
}

func UnmarshalServerConfigFromJson(bytes []byte, config *ServerConfig) error {
	return json.Unmarshal(bytes, config)
}
