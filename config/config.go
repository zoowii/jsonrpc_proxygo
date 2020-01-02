package config

import (
	"encoding/json"
)

type ServerConfig struct {
	Endpoint string `json:"endpoint"`

	// upstream plugin
	Plugins struct {
		Upstream struct {
			TargetEndpoints []string `json:"upstream_endpoints"`
		} `json:"upstream"`
	} `json:"plugins"`
}

func UnmarshalServerConfigFromJson(bytes []byte, config *ServerConfig) error {
	return json.Unmarshal(bytes, config)
}
