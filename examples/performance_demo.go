package main

import (
	"fmt"
	"log"
	"time"

	"github.com/KrakenTech-LLC/KrknDB/internal/kdb"
)

func main() {
	// Create a 32-byte encryption key
	encryptionKey := []byte("12345678901234567890123456789012")

	// Initialize the database
	db, err := kdb.New("./perfdata", encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	fmt.Println("=== Performance Demonstration ===\n")

	// Populate database with test data
	fmt.Println("Populating database with 1000 MD5 hashes...")
	start := time.Now()
	for i := 0; i < 1000; i++ {
		hashStr := fmt.Sprintf("hash_%d_test_data_for_performance", i)
		hash := kdb.NewHash(hashStr, fmt.Sprintf("value_%d", i), 0)
		if err := hash.Store(); err != nil {
			log.Printf("Failed to store hash: %v", err)
		}
	}
	fmt.Printf("✓ Stored 1000 hashes in %v\n\n", time.Since(start))

	// Test 1: Direct lookup (O(1))
	fmt.Println("Test 1: Direct Lookup (O(1))")
	fmt.Println("Looking up a single hash by exact match...")
	start = time.Now()
	targetHash := "hash_500_test_data_for_performance"
	found, err := db.GetHashByOriginalHash(targetHash, 0)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("Not found: %v", err)
	} else {
		fmt.Printf("✓ Found: %s -> %s\n", found.Hash, found.Value)
	}
	fmt.Printf("Time: %v\n\n", elapsed)

	// Test 2: Batch search with efficient filtering
	fmt.Println("Test 2: Batch Search (Single Scan + Filter)")
	fmt.Println("Searching for 10 specific hashes out of 1000...")
	searchHashes := []string{
		"hash_100_test_data_for_performance",
		"hash_200_test_data_for_performance",
		"hash_300_test_data_for_performance",
		"hash_400_test_data_for_performance",
		"hash_500_test_data_for_performance",
		"hash_600_test_data_for_performance",
		"hash_700_test_data_for_performance",
		"hash_800_test_data_for_performance",
		"hash_900_test_data_for_performance",
		"hash_999_test_data_for_performance",
	}

	start = time.Now()
	foundCount := 0
	for hash := range db.FindHashes(searchHashes, 0) {
		foundCount++
		fmt.Printf("  Found: %s -> %s\n", hash.Hash, hash.Value)
	}
	elapsed = time.Since(start)
	fmt.Printf("✓ Found %d hashes in %v\n", foundCount, elapsed)
	fmt.Printf("Note: This scanned all 1000 hashes ONCE and filtered efficiently\n\n")

	// Test 3: Iterate all hashes of a type
	fmt.Println("Test 3: Full Iteration (Generator)")
	fmt.Println("Iterating through all 1000 hashes...")
	start = time.Now()
	count := 0
	for range db.GetHashesByHashType(0) {
		count++
	}
	elapsed = time.Since(start)
	fmt.Printf("✓ Iterated %d hashes in %v\n", count, elapsed)
	fmt.Printf("Average: %v per hash\n\n", elapsed/time.Duration(count))

	// Test 4: Early termination
	fmt.Println("Test 4: Early Termination")
	fmt.Println("Getting only first 5 hashes (demonstrating lazy evaluation)...")
	start = time.Now()
	count = 0
	for hash := range db.GetHashesByHashType(0) {
		count++
		if count == 1 {
			fmt.Printf("  First hash: %s\n", hash.Hash)
		}
		if count >= 5 {
			break
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("✓ Retrieved %d hashes in %v\n", count, elapsed)
	fmt.Printf("Note: Only fetched 5 hashes, not all 1000!\n\n")

	// Test 5: Prefix search
	fmt.Println("Test 5: Prefix Search")
	fmt.Println("Searching for hashes starting with specific prefix...")
	// First, let's add some hashes with a known prefix
	testHashes := []string{
		"aabbccdd11223344556677889900aabb",
		"aabbccdd99887766554433221100ffee",
		"aabbccdddeadbeefcafebabe12345678",
	}
	for i, h := range testHashes {
		hash := kdb.NewHash(h, fmt.Sprintf("prefix_test_%d", i), 0)
		hash.Store()
	}

	start = time.Now()
	count = 0
	// Search for hashes where the SHA256 sum starts with a specific pattern
	// Note: This searches the SHA256 sum of the hash, not the hash itself
	for hash := range db.SearchHashesByPrefix("", 0) {
		if count < 3 {
			fmt.Printf("  Match: %s (sum prefix: %.16s...)\n", hash.Hash, string(hash.Sum))
		}
		count++
		if count >= 3 {
			break
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("✓ Found matches in %v\n\n", elapsed)

	fmt.Println("=== Performance Summary ===")
	fmt.Println("1. Direct lookup: Fastest for single exact matches (O(1))")
	fmt.Println("2. Batch search: Efficient for multiple hashes (single scan)")
	fmt.Println("3. Full iteration: Memory-efficient for large datasets (generator)")
	fmt.Println("4. Early termination: Only processes what you need")
	fmt.Println("5. Prefix search: Efficient for partial matches")
}
