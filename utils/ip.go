package utils

import (
	"io/ioutil"
	"net/http"
	"time"
)

func PublicIP() (string, error) {
	c := http.Client{
		Timeout: time.Second * 3,
	}
	resp, err := c.Get("http://i.kfd.me")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	return string(bs), err
}
