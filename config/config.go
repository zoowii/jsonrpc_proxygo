package config

import (
	"encoding/json"
	"strings"
)

// 本服务的配置信息
type ServerConfig struct {
	Resolver string `json:"resolver,omitempty"` // consul agent 配置
	ConfigFileResolver string `json:"config_file_resolver,omitempty"` // 加载整个config文件的consul kv http路径

	Endpoint string `json:"endpoint"`
	Provider string `json:"provider,omitempty"` // 'websocket', 'http', etc. default is 'websocket'

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

		Disable struct {
			Start bool `json:"start,omitempty"`
			DisabledRpcMethods []string `json:"disabled_rpc_methods"`
		} `json:"disable,omitempty"`

		RateLimit struct {
			Start bool `json:"start,omitempty"`
			ConnectionRate int `json:"connection_rate,omitempty"`
			RpcRate int `json:"rpc_rate,omitempty"`
		} `json:"rate_limit,omitempty"`
	} `json:"plugins,omitempty"`
}

// 从json中加载配置
func UnmarshalServerConfigFromJson(bytes []byte, config *ServerConfig) error {
	return json.Unmarshal(bytes, config)
}

// 判断是否是consul resolver url的格式
func IsConsulResolver(resolver string) bool {
	return len(resolver) > 0 && strings.Index(resolver, "consul://") == 0
}
