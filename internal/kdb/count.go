package kdb

import (
	"encoding/binary"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// PerformRecount recounts all hash types and updates all counters
// This is useful if counters get out of sync or corrupted
// Uses the hash type registry for efficient iteration
func (kc *KDB) PerformRecount() error {
	logger("Starting full recount of all hash types", Info)

	// Get all registered hash types
	hashTypes, err := kc.getRegisteredHashTypes()
	if err != nil {
		logger(fmt.Sprintf("Failed to get registered hash types: %v", err), Error)
		return fmt.Errorf("failed to get registered hash types: %w", err)
	}

	if len(hashTypes) == 0 {
		logger("No hash types registered, nothing to recount", Info)
		// Set total to 0
		if err := kc.setCount(totalHashesKey, 0); err != nil {
			return fmt.Errorf("failed to update total hash count: %w", err)
		}
		return nil
	}

	logger(fmt.Sprintf("Found %d registered hash types", len(hashTypes)), Info)

	totalCount := 0

	// Recount each registered hash type
	for _, hashType := range hashTypes {
		count := 0
		prefix := []byte(fmt.Sprintf(hashTypeLookupPrefix, hashType))

		// Count all hashes with this hash type prefix
		err := kc.c.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false // We only need to count keys
			opts.Prefix = prefix

			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				count++
			}

			return nil
		})

		if err != nil {
			logger(fmt.Sprintf("Failed to count hash type %d: %v", hashType, err), Error)
			return fmt.Errorf("failed to count hash type %d: %w", hashType, err)
		}

		// Update the counter for this hash type
		countKey := fmt.Sprintf(hashTypeCountPrefix, hashType)
		if err := kc.setCount(countKey, count); err != nil {
			logger(fmt.Sprintf("Failed to update count for hash type %d: %v", hashType, err), Error)
			return fmt.Errorf("failed to update count for hash type %d: %w", hashType, err)
		}

		logger(fmt.Sprintf("Updated hash type %d: %d hashes", hashType, count), Info)
		totalCount += count
	}

	// Update the total hash count
	if err := kc.setCount(totalHashesKey, totalCount); err != nil {
		logger(fmt.Sprintf("Failed to update total hash count: %v", err), Error)
		return fmt.Errorf("failed to update total hash count: %w", err)
	}

	logger(fmt.Sprintf("Recount completed successfully: %d total hashes across %d hash types", totalCount, len(hashTypes)), Info)
	return nil
}

// RecountHashType recounts hashes for a specific hash type and updates its counter
func (kc *KDB) RecountHashType(hashType uint64) error {
	logger(fmt.Sprintf("Starting recount for hash type %d", hashType), Info)

	count := 0
	prefix := []byte(fmt.Sprintf(hashTypeLookupPrefix, hashType))

	// Count all hashes with this hash type prefix
	err := kc.c.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need to count keys
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}

		return nil
	})

	if err != nil {
		logger(fmt.Sprintf("Failed to count hash type %d: %v", hashType, err), Error)
		return fmt.Errorf("failed to count hash type %d: %w", hashType, err)
	}

	// Update the counter for this hash type
	countKey := fmt.Sprintf(hashTypeCountPrefix, hashType)
	if err := kc.setCount(countKey, count); err != nil {
		logger(fmt.Sprintf("Failed to update count for hash type %d: %v", hashType, err), Error)
		return fmt.Errorf("failed to update count for hash type %d: %w", hashType, err)
	}

	logger(fmt.Sprintf("Recount for hash type %d completed: %d hashes", hashType, count), Info)
	return nil
}

// setCount sets a counter to a specific value (used by recount operations)
func (kc *KDB) setCount(key string, count int) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	return kc.c.Update(func(txn *badger.Txn) error {
		// Store as binary uint64 (8 bytes)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(count))
		return txn.Set([]byte(key), buf)
	})
}
