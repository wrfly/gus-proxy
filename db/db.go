package db

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/boltdb/bolt"
)

// DNS DNS database
type DNS struct {
	db *bolt.DB
}

// Open a database file,create one if not exist
func (d *DNS) Open() error {
	dbFileName := path.Join(os.TempDir(), "gus-dns.db")
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		// create the file
		_, err := os.Create(dbFileName)
		if err != nil {
			return err
		}
	}
	db, err := bolt.Open(dbFileName, 0600, nil)
	if err != nil {
		return err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("DNS"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	d.db = db
	return nil
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
		b := tx.Bucket([]byte("DNS"))
		err := b.Put([]byte(domain), []byte(answerStr))
		return err
	})
	return err
}

// Query domain from DB
func (d *DNS) Query(domain string) (answer []string) {
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DNS"))
		v := b.Get([]byte(domain))
		answer = strings.Split(string(v), "|")
		return nil
	})
	if answer[0] == "" {
		return nil
	}
	return answer
}
