package utils

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

func HashSlice(slice []string) string {
	sort.Strings(slice)
	s := strings.Join(slice, "")
	return fmt.Sprintf("%x", md5.New().Sum([]byte(s)))
}
