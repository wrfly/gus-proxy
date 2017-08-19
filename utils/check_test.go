package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wrfly/gus-proxy/types"
)

func TestCheckProxyAvailable(t *testing.T) {
	host := types.ProxyHost{
		// Addr: "http://61.130.97.212:8099", // www.baidu.com
		Addr: "http://119.75.216.20:80", // www.baidu.com
	}
	err := CheckProxyAvailable(host)
	assert.Error(t, err)
}
