package common

import (
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

func GetSessionStringParam(session *proxy.JSONRpcRequestSession, paramName string, defaultValue *string) (result string, err error) {
	paramValue, ok := session.Parameters[paramName]
	var defaultValueStr = ""
	if defaultValue != nil {
		defaultValueStr = *defaultValue
	}
	if !ok {
		result = defaultValueStr
		return
	}
	result, ok = paramValue.(string)
	if !ok {
		err = errors.New("invalid " + paramName + " param in session")
		return
	}
	return
}

func SetSelectedUpstreamTargetEndpoint(session *proxy.ConnectionSession, value string) (err error) {
	session.SelectedUpstreamTarget = &value
	return
}

func GetSelectedUpstreamTargetEndpoint(session *proxy.ConnectionSession, defaultValue *string) (result string, err error) {
	var defaultValueStr = ""
	if defaultValue != nil {
		defaultValueStr = *defaultValue
	}
	if session.SelectedUpstreamTarget == nil {
		result = defaultValueStr
		return
	}
	result = *session.SelectedUpstreamTarget
	return
}