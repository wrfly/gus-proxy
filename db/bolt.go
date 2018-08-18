package db

import (
	"fmt"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"github.com/wrfly/go/src/math/rand"
	"github.com/wrfly/gus-proxy/utils"
)

var (
	bktName = []byte("DNS")
)

// DNS database
type DNS struct {
	db *bolt.DB
}

// New database file, create one if not exist
func New(dbFileName string) (*DNS, error) {
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		// create the file
		_, err := os.Create(dbFileName)
		if err != nil {
			return nil, err
		}
	}
	dnsDB, err := bolt.Open(dbFileName, 0600, nil)
	if err != nil {
		return nil, err
	}

	dnsDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(bktName)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &DNS{
		db: dnsDB,
	}, nil
}

// Close the DB
func (d *DNS) Close() error {
	path := d.db.Path()
	err := d.db.Close()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// setDNS to the DB
func (d *DNS) setDNS(domain string, answer []string) error {
	answerStr := strings.Join(answer, "|")
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bktName)
		err := b.Put([]byte(domain), []byte(answerStr))
		return err
	})
	return err
}

// query domain from DB
func (d *DNS) query(domain string) (answer []string) {
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bktName)
		v := b.Get([]byte(domain))
		answer = strings.Split(string(v), "|")
		return nil
	})
	if len(answer) == 0 {
		return nil
	}
	if answer[0] == "" {
		return nil
	}
	return answer
}

// SelectIP domain -> IP
// kfd.me:8080 -> xx.x.xx.x:8080
func (dnsDB *DNS) SelectIP(host string) string {
	str := strings.Split(host, ":")
	domain := str[0]
	port := "80"
	if len(str) == 2 {
		port = str[1]
	}

	ips := dnsDB.query(domain)
	logrus.Debugf("query %s IP: %v", domain, ips)
	// not found in db
	if len(ips) == 0 {
		digIPs, err := utils.LookupHost(domain)
		if err != nil {
			logrus.Errorf("Dig Error: %s", err)
			return "127.0.0.1:80"
		}
		// set to db
		logrus.Debugf("Set DNS DB: domain: %s IP: %v", domain, digIPs)
		if err := dnsDB.setDNS(domain, digIPs); err != nil {
			logrus.Error(err)
		}
		ips = digIPs
	}

	ip := ips[rand.Int()%len(ips)]
	return fmt.Sprintf("%s:%s", ip, port)
}
