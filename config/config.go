package config

import (
	"encoding/json"
	"strconv"
	"strings"
)

// consul相关的配置信息
type ConsulConfig struct {
	Endpoint                   string   `json:"endpoint,omitempty"`      // consul agent的endpoint
	Id                         string   `json:"id,omitempty"`            // 服务id
	Name                       string   `json:"name,omitempty"`          // 服务名称
	Tags                       []string `json:"tags,omitempty"`          // 注册服务的标签
	HealthCheckIntervalSeconds int      `json:"health_checker_interval"` // consul服务心跳上报的间隔秒数

	ConfigFileResolver string `json:"config_file_resolver,omitempty"` // 加载整个config文件的consul kv http路径

	HealthCheckId string // 心跳检查的Check ID
}

// 本服务的配置信息
type ServerConfig struct {
	Resolver *ConsulConfig `json:"resolver,omitempty"` // consul agent配置

	Endpoint string `json:"endpoint"`
	Provider string `json:"provider,omitempty"` // 'websocket', 'http', etc. default is 'websocket'

	Log struct {
		Level      string `json:"level,omitempty"` // DEBUG,INFO,WARN,ERROR, INFO is default
		OutputFile string `json:"output_file,omitempty"`
	} `json:"log,omitempty"`

	Plugins struct {
		// upstream plugin config
		Upstream struct {
			TargetEndpoints []struct {
				Url    string `json:"url"`
				Weight int64  `json:"weight"`
			} `json:"upstream_endpoints"`
		} `json:"upstream,omitempty"`

		// cache plugin config
		Caches []struct {
			Name           string        `json:"name"`
			ParamsForCache []interface{} `json:"paramsForCache"`
			ExpireSeconds  int64         `json:"expire_seconds"`
		} `json:"caches,omitempty"`

		BeforeCacheConfigs []struct {
			MethodName                   string `json:"method"`
			FetchCacheKeyFromParamsCount int    `json:"fetch_cache_key_from_params_count"`
		} `json:"before_cache_configs,omitempty"`

		Statistic struct {
			Start bool `json:"start,omitempty"`
		} `json:"statistic,omitempty"`

		Disable struct {
			Start              bool     `json:"start,omitempty"`
			DisabledRpcMethods []string `json:"disabled_rpc_methods"`
		} `json:"disable,omitempty"`

		RateLimit struct {
			Start          bool `json:"start,omitempty"`
			ConnectionRate int  `json:"connection_rate,omitempty"`
			RpcRate        int  `json:"rpc_rate,omitempty"`
		} `json:"rate_limit,omitempty"`
	} `json:"plugins,omitempty"`
}

func (serverConfig ServerConfig) GetEndpointPort() *int {
	endpoint := serverConfig.Endpoint
	if len(endpoint) < 1 {
		return nil
	}
	colonPos := strings.LastIndex(endpoint, ":")
	if colonPos < 0 {
		return nil
	}
	portStr := endpoint[colonPos+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil
	}
	return &port
}

func (serverConfig ServerConfig) GetEndpointHost() *string {
	endpoint := serverConfig.Endpoint
	if len(endpoint) < 1 {
		return nil
	}
	colonPos := strings.LastIndex(endpoint, ":")
	if colonPos < 0 {
		return nil
	}
	hostname := endpoint[0:colonPos]
	return &hostname
}

// 从json中加载配置
func UnmarshalServerConfigFromJson(bytes []byte, config *ServerConfig) error {
	return json.Unmarshal(bytes, config)
}

// 判断是否是consul resolver url的格式
func IsConsulResolver(resolver string) bool {
	return len(resolver) > 0 && strings.Index(resolver, "consul://") == 0
}
