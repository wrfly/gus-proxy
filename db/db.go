package db

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/boltdb/bolt"
)

var (
	bktName = []byte("DNS")
)

// DNS database
type DNS struct {
	db *bolt.DB
}

// New database file, create one if not exist
func New() (*DNS, error) {
	dbFileName := path.Join(os.TempDir(), "gus-dns.db")
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

// SetDNS to the DB
func (d *DNS) SetDNS(domain string, answer []string) error {
	answerStr := strings.Join(answer, "|")
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bktName)
		err := b.Put([]byte(domain), []byte(answerStr))
		return err
	})
	return err
}

// Query domain from DB
func (d *DNS) Query(domain string) (answer []string) {
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
