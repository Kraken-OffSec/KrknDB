# KrknDB - Fast Hash Lookup Database

A high-performance key-value database optimized for storing and retrieving cryptographic hashes with their cracked values. Built on BadgerDB with AES-256 encryption.

## Features

✅ **Fast Lookups**: O(1) direct hash lookups  
✅ **Efficient Batch Search**: Single-scan filtering for multiple hashes  
✅ **Memory Efficient**: Generator-based iteration for large datasets  
✅ **Type-based Organization**: Group hashes by algorithm (MD5, SHA256, etc.)  
✅ **Encrypted Storage**: AES-256 encryption at rest  
✅ **Thread Safe**: Concurrent access with mutex protection  
✅ **Go 1.23+ Iterators**: Modern, idiomatic Go API  

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/KrakenTech-LLC/KrknDB/internal/kdb"
)

func main() {
    // Initialize database with 32-byte encryption key
    encryptionKey := []byte("12345678901234567890123456789012")
    db, _ := kdb.New("./data", encryptionKey)
    defer db.Close()

    // Store a hash
    hash := kdb.NewHash("5f4dcc3b5aa765d61d8327deb882cf99", "password", 0)
    hash.Store()

    // Find it
    found, _ := db.GetHashByOriginalHash("5f4dcc3b5aa765d61d8327deb882cf99", 0)
    fmt.Printf("Cracked: %s\n", found.Value) // Output: Cracked: password
}
```

## Key Format

Hashes are stored with the key format:
```
krkn:{hashType}:{hexEncodedSHA256Sum}
```

Examples:
- `krkn:0:5f4dcc3b...` (MD5, type 0)
- `krkn:1400:5e8848...` (SHA256, type 1400)
- `krkn:1700:b109f3...` (SHA512, type 1700)

## API Methods

### Store Hash
```go
hash := kdb.NewHash("5f4dcc3b5aa765d61d8327deb882cf99", "password", 0)
err := hash.Store()
```

### Direct Lookup (Single Hash) - O(1)
```go
hash, err := db.GetHashByOriginalHash("5f4dcc3b5aa765d61d8327deb882cf99", 0)
```
**Best for:** Finding 1-5 specific hashes

### Batch Search (Multiple Hashes) - O(m)
```go
searchList := []string{"hash1", "hash2", "hash3", ...}
for hash := range db.FindHashesByHashSum(searchList, 0) {
    fmt.Printf("Found: %s -> %s\n", hash.Hash, hash.Value)
}
```
**Best for:** Finding 10+ hashes efficiently  
**Performance:** Single database scan with O(1) hashmap filtering

### Iterate by Type - O(m)
```go
for hash := range db.GetHashesByHashType(0) {
    fmt.Printf("%s -> %s\n", hash.Hash, hash.Value)
    if someCondition {
        break // Early termination supported
    }
}
```
**Best for:** Processing all hashes of a specific algorithm

### Prefix Search - O(k)
```go
for hash := range db.SearchHashesByPrefix("5f4d", 0) {
    fmt.Printf("Match: %s\n", hash.Hash)
}
```
**Best for:** Partial hash lookups, autocomplete

## Performance

### Batch Search Efficiency

**Scenario:** 1M total hashes, 100k MD5 hashes, searching for 1000 specific hashes

| Method | Operations | Database Scans | Winner |
|--------|-----------|----------------|--------|
| Individual Lookups | ~20,000 | 1,000 | ❌ |
| Single Scan + Filter | ~100,000 | 1 | ✅ |

The single-scan approach becomes MORE efficient as the number of search hashes increases.

### Memory Efficiency

Generator pattern loads only one hash at a time:
```go
// Can iterate millions of hashes without OOM
for hash := range db.GetHashesByHashType(0) {
    // Only one hash in memory at a time
}
```

## Hash Types (Hashcat Codes)

Common hash type codes:
- `0` - MD5
- `100` - SHA1
- `1400` - SHA256
- `1700` - SHA512
- `1000` - NTLM
- `3000` - LM
- `5600` - NetNTLMv2

See [Hashcat documentation](https://hashcat.net/wiki/doku.php?id=example_hashes) for complete list.

## Examples

### Basic Usage
```bash
cd examples
go run basic_usage.go
```

### Performance Demo
```bash
cd examples
go run performance_demo.go
```

## Documentation

- **FIXES_SUMMARY.md** - What was wrong and how it was fixed
- **ARCHITECTURE.md** - Detailed architecture and design decisions
- **examples/** - Working code examples

## Requirements

- Go 1.23+ (for `iter.Seq` support)
- BadgerDB v3

## Thread Safety

All operations are thread-safe and can be called concurrently from multiple goroutines.

## Use Cases

- Password cracking databases
- Rainbow tables
- Hash breach databases
- Security research
- Cryptographic hash analysis

## License

See LICENSE file for details.

---

**Built with ❤️ for fast hash lookups**

