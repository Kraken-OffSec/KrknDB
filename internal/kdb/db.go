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

// KDB represents the key-value database
type KDB struct {
	encryptionKey []byte
	c             *badger.DB // badger database
	mu            sync.Mutex // mutex for concurrent access
	isNew         bool       // true if the cache is new
	absPath       string     // absolute path to the database file
	parentFolder  string     // absolute path to the parent folder
}

// New creates a new KDB instance
// dbFolder is the folder to store the database in
// encryptionKey is the 32-byte encryption key
// opts is an optional set of KDBOptions
func New(dbFolder string, encryptionKey []byte, opts ...*Options) (*KDB, error) {
	var (
		err       error
		absPath   string
		dbOptions *Options
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

	if len(opts) > 0 {
		dbOptions = DefaultOptions()
		dbOptions.ValueDir = absPath
	} else {
		dbOptions = opts[0]
	}

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
			WithValueDir(absPath).                                                      // Use the same directory for data and value files
			WithEncryptionKey(encryptionKey).                                           // Enable encryption
			WithCompression(dbOptions.Compression).                                     // Use ZSTD compression
			WithEncryptionKeyRotationDuration(dbOptions.EncryptionKeyRotationDuration). // Rotate keys daily
			WithNumVersionsToKeep(dbOptions.NumVersionsToKeep).                         // Only keep the latest version of each key
			// WithBlockCacheSize(8 << 30).                       						// 8GB block cache
			WithIndexCacheSize(dbOptions.IndexCacheSize).                   // 10GB index cache
			WithValueThreshold(dbOptions.ValueThreshold).                   // 64KB inline threshold
			WithValueLogFileSize(dbOptions.ValueLogFileSize).               // 2GB log files
			WithMemTableSize(dbOptions.MemTableSize).                       // 512MB memtables
			WithNumMemtables(dbOptions.NumMemTables).                       // More in-RAM tables
			WithNumCompactors(dbOptions.NumCompactors).                     // More compaction threads
			WithNumLevelZeroTables(dbOptions.NumLevelZeroTables).           // 20 L0 tables before compaction
			WithNumLevelZeroTablesStall(dbOptions.NumLevelZeroTablesStall). // 40 L0 tables before stalling
			WithBaseLevelSize(dbOptions.BaseLevelSize).                     // 20GB base level
			WithMaxLevels(dbOptions.MaxLevels).                             // 7 levels
			WithBloomFalsePositive(dbOptions.BloomFalsePositive).           // 1% false positive rate
			WithLogger(nil)                                                 // Disable logging for speed

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

// Get returns the global KDB instance
func Get() *KDB {
	return cache
}

// IsNew returns true if this is a freshly created database
func (kc *KDB) IsNew() bool {
	return kc.isNew
}

// Close closes the database
func (kc *KDB) Close() error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	return kc.c.Close()
}

// Nil returns true if the database is nil
func (kc *KDB) Nil() bool {
	return kc.c == nil
}

// ParentFolder returns the parent folder of the database
func (kc *KDB) ParentFolder() string {
	return kc.parentFolder
}

// DBPath returns the absolute path to the database file
func (kc *KDB) DBPath() string {
	return kc.absPath
}
