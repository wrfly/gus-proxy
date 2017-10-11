package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB(t *testing.T) {
	dnsDB, err := New()
	assert.NoError(t, err)
	defer dnsDB.Close()

	err = dnsDB.SetDNS("kfd.me", []string{"8.8.8.8", "1.1.1.2"})
	assert.NoError(t, err)

	a := dnsDB.Query("kfd.me")
	for _, ip := range a {
		fmt.Println(ip)
	}

	b := dnsDB.Query("kfd.mee")
	for _, ip := range b {
		fmt.Println(ip)
	}
	fmt.Println("done")
}
