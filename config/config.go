package config

import (
	"encoding/json"
)

type ServerConfig struct {
	Endpoint string `json:"endpoint"`

	Plugins struct {
		// upstream plugin config
		Upstream struct {
			TargetEndpoints []string `json:"upstream_endpoints"`
		} `json:"upstream"`

		// cache plugin config
		Caches []struct {
			Name string `json:"name"`
			ExpireSeconds int64 `json:"expire_seconds"`
		} `json:"caches"`

		BeforeCacheConfigs []struct {
			MethodName string `json:"method"`
			FetchCacheKeyFromParamsCount int `json:"fetch_cache_key_from_params_count"`
		} `json:"before_cache_configs"`
	} `json:"plugins"`
}

func UnmarshalServerConfigFromJson(bytes []byte, config *ServerConfig) error {
	return json.Unmarshal(bytes, config)
}
