package ss

import (
	"encoding/base64"
	"flag"
	"net/url"
	"strings"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	log "github.com/sirupsen/logrus"
)

var config struct {
	Verbose    bool
	UDPTimeout time.Duration
	TCPCork    bool
}

func load() {
	var flags struct {
		Client   string
		Server   string
		Cipher   string
		Key      string
		Password string
		Keygen   int
		Socks    string
	}

	flag.StringVar(&flags.Cipher, "cipher", "AEAD_CHACHA20_POLY1305", "available ciphers: "+strings.Join(core.ListCipher(), " "))
	flag.StringVar(&flags.Password, "password", "", "password")
	flag.StringVar(&flags.Key, "key", "", "key")

	var key []byte
	if flags.Key != "" {
		k, err := base64.URLEncoding.DecodeString(flags.Key)
		if err != nil {
			log.Fatal(err)
		}
		key = k
	}

	addr := flags.Client
	cipher := flags.Cipher
	password := flags.Password
	var err error

	if strings.HasPrefix(addr, "ss://") {
		addr, cipher, password, err = parseURL(addr)
		if err != nil {
			log.Fatal(err)
		}
	}

	ciph, err := core.PickCipher(cipher, key, password)
	if err != nil {
		log.Fatal(err)
	}

	socksLocal(flags.Socks, addr, ciph.StreamConn)
}

func parseURL(s string) (addr, cipher, password string, err error) {
	u, err := url.Parse(s)
	if err != nil {
		return
	}

	addr = u.Host
	if u.User != nil {
		cipher = u.User.Username()
		password, _ = u.User.Password()
	}
	return
}
