package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/types"
)

func TestConfig(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c := Config{
		Debug:         true,
		ProxyFilePath: "../proxyhosts.txt",
		Scheduler:     types.PING,
		ListenPort:    "8088",
	}

	t.Run("validate config test", func(t *testing.T) {
		if err := c.Validate(); err != nil {
			t.Error(err)
		}
	})

	t.Run("mock proxy not found error", func(t *testing.T) {
		ori := c.ProxyFilePath
		c.ProxyFilePath = "https://kfd.me/404"
		err := c.Validate()
		assert.Error(t, err, "not found")
		c.ProxyFilePath = ori
	})

	t.Run("load hostfile", func(t *testing.T) {
		err := c.loadHosts()
		if err != nil {
			t.Error(err)
		}
		for _, host := range c.proxyHosts {
			t.Logf("got proxy: %s", host.Addr)
		}
	})

	t.Run("update host", func(t *testing.T) {
		err := c.loadHosts()
		if err != nil {
			t.Error(err)
		}
		c.UpdateProxies()
		for _, host := range c.availableProxyHosts {
			t.Logf("got available proxy: %s", host.Addr)
		}
	})
}
