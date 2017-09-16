package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/boltdb/bolt"
)

type DNS struct {
	db *bolt.DB
}

func (d *DNS) Open() error {
	f, err := ioutil.TempFile(os.TempDir(), "gus-dns-")
	if err != nil {
		return err
	}
	db, err := bolt.Open(f.Name(), 0600, nil)
	if err != nil {
		return err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	d.db = db
	return nil
}

func (d *DNS) Close() error {
	path := d.db.Path()
	err := d.db.Close()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (d *DNS) SetDNS(domain string, answer []string) error {
	answerStr := strings.Join(answer, "|")
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DNS"))
		err := b.Put([]byte(domain), []byte(answerStr))
		return err
	})
	return err
}

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
