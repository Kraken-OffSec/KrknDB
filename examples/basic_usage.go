package main

import (
	"fmt"
	"log"

	"github.com/KrakenTech-LLC/KrknDB/internal/kdb"
)

func main() {
	// Create a 32-byte encryption key (in production, use a secure key)
	encryptionKey := []byte("12345678901234567890123456789012")

	// Initialize the database
	db, err := kdb.New("./data", encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database initialized successfully")

	// Example 1: Store some hashes
	fmt.Println("\n=== Storing Hashes ===")
	hashes := []struct {
		hash     string
		value    string
		hashType uint64
	}{
		{"5f4dcc3b5aa765d61d8327deb882cf99", "password", 0},     // MD5
		{"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8", "password", 1400}, // SHA256
		{"b109f3bbbc244eb82441917ed06d618b9008dd09b3befd1b5e07394c706a8bb980b1d7785e5976ec049b46df5f1326af5a2ea6d103fd07c95385ffab0cacbc86", "password", 1700}, // SHA512
		{"482c811da5d5b4bc6d497ffa98491e38", "password123", 0}, // MD5
	}

	for _, h := range hashes {
		hash := kdb.NewHash(h.hash, h.value, h.hashType)
		if err := hash.Store(); err != nil {
			log.Printf("Failed to store hash: %v", err)
		} else {
			fmt.Printf("Stored hash: %s (type: %d)\n", h.hash, h.hashType)
		}
	}

	// Example 2: Retrieve hashes by hash type using generator
	fmt.Println("\n=== Retrieving MD5 Hashes (type 0) ===")
	count := 0
	for hash := range db.GetHashesByHashType(0) {
		count++
		fmt.Printf("Hash #%d: %s -> %s (sum: %s)\n", count, hash.Hash, hash.Value, string(hash.Sum))
	}

	// Example 3: Retrieve SHA256 hashes (type 1400)
	fmt.Println("\n=== Retrieving SHA256 Hashes (type 1400) ===")
	count = 0
	for hash := range db.GetHashesByHashType(1400) {
		count++
		fmt.Printf("Hash #%d: %s -> %s\n", count, hash.Hash, hash.Value)
	}

	// Example 4: Find a specific hash by its sum
	fmt.Println("\n=== Finding Hash by Sum ===")
	targetHash := "5f4dcc3b5aa765d61d8327deb882cf99"
	foundHash, err := db.GetHashBySum(
		string(kdb.NewHash(targetHash, "", 0).Sum),
		0,
	)
	if err != nil {
		log.Printf("Hash not found: %v", err)
	} else {
		fmt.Printf("Found: %s -> %s\n", foundHash.Hash, foundHash.Value)
	}

	// Example 5: Search by prefix
	fmt.Println("\n=== Searching by Prefix ===")
	// Search for hashes starting with "5f" in MD5 (type 0)
	count = 0
	for hash := range db.SearchHashesByPrefix("5f", 0) {
		count++
		fmt.Printf("Match #%d: %s -> %s\n", count, hash.Hash, hash.Value)
	}

	// Example 6: Demonstrate early termination of generator
	fmt.Println("\n=== Early Termination (first 2 MD5 hashes only) ===")
	count = 0
	for hash := range db.GetHashesByHashType(0) {
		count++
		fmt.Printf("Hash #%d: %s -> %s\n", count, hash.Hash, hash.Value)
		if count >= 2 {
			break // This will stop the iteration
		}
	}

	fmt.Println("\n=== Done ===")
}

