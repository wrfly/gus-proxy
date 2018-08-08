package utils

import (
	"os/exec"
	"strconv"
	"strings"
)

// Ping host and returns the average rtt
func Ping(ip string) float32 {
	cmd := exec.Command("ping", "-A", "-c", "3", "-w", "2", ip)
	b, err := cmd.Output()
	if err != nil {
		println(err.Error())
		return 9999
	}

	o := string(b)
	l := strings.LastIndex(o, "=")
	o = string(o[l+2:])
	s := strings.Split(o, "/")

	avg := s[1]

	f, _ := strconv.ParseFloat(avg, 32)

	return float32(f)
}
