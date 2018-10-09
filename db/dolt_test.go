package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB(t *testing.T) {
	dnsDB, err := New("/tmp/gus.db")
	assert.NoError(t, err)
	defer dnsDB.Close()

	err = dnsDB.setDNS("kfd.me", []string{"8.8.8.8", "1.1.1.2"})
	assert.NoError(t, err)

	t.Run("query no error", func(t *testing.T) {
		a := dnsDB.query("kfd.me")
		for _, ip := range a {
			fmt.Println(ip)
		}
	})

	t.Run("query not found", func(t *testing.T) {
		b := dnsDB.query("kfd.mee")
		if len(b) != 0 {
			t.Error("wtf?")
		}
	})

}
