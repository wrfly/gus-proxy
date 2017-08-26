package config

import (
	"fmt"
	"testing"
)

func TestLoadHostFile(t *testing.T) {
	c := Config{
		ProxyHostsFile: "../proxyhosts.txt",
	}
	hosts, err := c.LoadHosts()
	if err != nil {
		t.Error(err)
	}
	for _, host := range hosts {
		fmt.Println(host.Addr)
	}
}
