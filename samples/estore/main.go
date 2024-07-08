package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/estore"
)

const keyFile = "/db/sealed_key"

func main() {
	// Open existing DB or create a new one
	var db *estore.DB
	sealedKey, err := os.ReadFile(keyFile)
	if err == nil {
		fmt.Println("Found existing DB")
		db, err = openExistingDB(sealedKey)
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Creating new DB")
		db, err = createNewDB()
	}

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get the value of the key
	value, closer, err := db.Get([]byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()
	fmt.Printf("hello=%s\n", value)
}

func createNewDB() (*estore.DB, error) {
	// Generate an encryption key
	encryptionKey := make([]byte, 16)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Seal the encryption key
	sealedKey, err := ecrypto.SealWithUniqueKey(encryptionKey, nil)
	if err != nil {
		return nil, err
	}
	if err := os.Mkdir("/db", 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyFile, sealedKey, 0o600); err != nil {
		return nil, err
	}

	// Create an encrypted store
	opts := &estore.Options{
		EncryptionKey: encryptionKey,
	}
	db, err := estore.Open("/db", opts)
	if err != nil {
		return nil, err
	}

	// Set a key-value pair
	if err := db.Set([]byte("hello"), []byte("world"), nil); err != nil {
		return nil, err
	}

	return db, nil
}

func openExistingDB(sealedKey []byte) (*estore.DB, error) {
	encryptionKey, err := ecrypto.Unseal(sealedKey, nil)
	if err != nil {
		return nil, err
	}
	opts := &estore.Options{
		EncryptionKey: encryptionKey,
	}
	return estore.Open("/db", opts)
}
