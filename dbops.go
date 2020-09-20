// Functions to process db operations
package main

import (
	"fmt"
	"log"	
	"time"
	"github.com/boltdb/bolt"
)

var db *bolt.DB

// create and initialize db session
func createDb() {
	// Init the db
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	d, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})	
	if err != nil {
		log.Fatal(err)
	}
	db = d

	// create bucket for users
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Users"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	// insert default username
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		err := b.Put([]byte("admin"), []byte("password"))
		return err
	})

	// create bucket for sessions
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Sessions"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})	
}

func getStringFromDB(bucketName string, key string) string {
	// get from db
	value, _ := queryDB([]byte(bucketName), []byte(key))
	return string(value)
}

func queryDB(bn, k []byte) (val []byte, length int) {
	err := db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bn)
		if bkt == nil {
			return fmt.Errorf("bucket %q not found", bn)
		}
		val = bkt.Get(k)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return val, len(string(val))
}

func updateDBString(bucketName, key, value string) {
	updateDB([]byte(bucketName), []byte(key), []byte(value))
}

func updateDB(bucketName, key, value []byte) {
	err := db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		err = bkt.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func deleteKeyString(bucketName, key string) {
	deleteKey([]byte(bucketName), []byte(key))
}

func deleteKey(bucketName, keyName []byte) {
	err := db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket(bucketName)
		err := b.Delete(keyName)

		return err
	})

	if err != nil {
		log.Fatalf("failure : %s\n", err)
	}
}
