// +build !windows

package registry

import "fmt"

func SetGlobalProxy(port int, config *ProxyConfig) error {
	if port > 0 {
		return nil
	}
	return fmt.Errorf("invalid socks port %d", port)
}

func CleanGlobalProxy(config *ProxyConfig) {
}

func SetHttpProxyEnvironmentVariable(port int, config *ProxyConfig) error {
	return nil
}

func CleanHttpProxyEnvironmentVariable(config *ProxyConfig) {

}

func ResetGlobalProxyAndEnvironmentVariable() {

}
