package kdb

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/KrakenTech-LLC/KrknDB/internal/util"
	"github.com/dgraph-io/badger/v3"
)

var (
	initOnce sync.Once
	cache    *KDB
)

const (
	storedHashPrefix     = "krkn:%d:%v" // hash_type:stored_hash.Key
	hashTypeLookupPrefix = "krkn:%d"    // hash_type
)

type KDB struct {
	encryptionKey []byte
	c             *badger.DB // badger database
	mu            sync.Mutex // mutex for concurrent access
	isNew         bool       // true if the cache is new
	absPath       string     // absolute path to the database file
	parentFolder  string     // absolute path to the parent folder
}

func New(dbFolder string, encryptionKey []byte) (*KDB, error) {
	var (
		err     error
		absPath string
	)

	if len(encryptionKey) == 0 || len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	// Get the absolute path for the parent folder
	absPath, err = filepath.Abs(dbFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for '%s': %w", dbFolder, err)
	}

	// Get the absolute path for the database file
	dbFile := filepath.Join(absPath, "krkn.db")

	initOnce.Do(func() {
		// Check if the krkn database already exists
		isNewDB := !util.PathExists(absPath)

		if isNewDB {
			// Create the krkn database directory
			if err = os.MkdirAll(absPath, 0700); err != nil {
				err = fmt.Errorf("failed to create krkn database directory: %w", err)
				return
			}
		}

		// Configure BadgerDB options
		opts := badger.DefaultOptions(absPath).
			WithEncryptionKey(encryptionKey).
			WithIndexCacheSize(10 << 20).
			WithLoggingLevel(badger.ERROR)

		// Try to open the database with retries
		var db *badger.DB
		maxRetries := 3
		retryDelay := 3 * time.Second
		for i := 0; i < maxRetries; i++ {
			db, err = badger.Open(opts)
			if err == nil {
				// Successfully opened
				break
			}

			if i < maxRetries-1 {
				time.Sleep(retryDelay)
			}
		}

		if err != nil {
			err = fmt.Errorf("failed to open krkn database after %d retries: %w", maxRetries, err)
			return
		}

		cache = &KDB{
			encryptionKey: encryptionKey,
			c:             db,
			mu:            sync.Mutex{},
			isNew:         isNewDB,
			absPath:       dbFile,
			parentFolder:  absPath,
		}
	})

	if err != nil {
		return nil, err
	}

	return cache, nil
}

func Get() *KDB {
	return cache
}

// IsNew returns true if this is a freshly created database
func (kc *KDB) IsNew() bool {
	return kc.isNew
}

func (kc *KDB) Close() error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	return kc.c.Close()
}

func (kc *KDB) Nil() bool {
	return kc.c == nil
}

func (kc *KDB) ParentFolder() string {
	return kc.parentFolder
}

func (kc *KDB) DBPath() string {
	return kc.absPath
}
