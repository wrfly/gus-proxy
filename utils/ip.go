package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

func PublicIP() (string, error) {
	c := http.Client{
		Timeout: time.Second * 3,
	}
	resp, err := c.Get("http://ipinfo.io")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	x := &IPinfoJson{}
	err = json.Unmarshal(bs, x)
	return x.IP, err
}
