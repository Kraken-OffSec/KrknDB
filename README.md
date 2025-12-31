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


## Query Methods

### 1. Find Single Hash (Fastest)
```go
// O(1) - Direct lookup
hash, err := db.GetHashByOriginalHash("5f4dcc3b5aa765d61d8327deb882cf99", 0)
```
**Use when:** You know the exact hash you're looking for

### 2. Find Multiple Hashes (Efficient Batch)
```go
// O(m) - Single scan + filter
searchList := []string{"hash1", "hash2", "hash3"}
for hash := range db.FindHashesByHashSum(searchList, 0) {
    // Process found hash
}
```
**Use when:** Searching for 10+ hashes at once

### 3. Iterate All of Type
```go
// O(m) - Full iteration with early termination
for hash := range db.GetHashesByHashType(0) {
    // Process each hash
    if condition {
        break // Stop early
    }
}
```
**Use when:** Processing all hashes of a specific algorithm

### 4. Prefix Search
```go
// O(k) - Prefix scan
for hash := range db.SearchHashesByPrefix("5f4d", 0) {
    // Process matching hash
}
```
**Use when:** Partial hash lookups

## Decision Tree

```
How many hashes to find?
│
├─ 1 hash
│  └─ Use: GetHashByOriginalHash()
│     Complexity: O(1)
│
├─ 2-10 hashes
│  └─ Use: Either method works
│     Complexity: O(1) to O(m)
│
├─ 10+ hashes
│  └─ Use: FindHashesByHashSum()
│     Complexity: O(m) single scan
│
├─ All hashes of type
│  └─ Use: GetHashesByHashType()
│     Complexity: O(m) with early termination
│
└─ Partial match
   └─ Use: SearchHashesByPrefix()
      Complexity: O(k)
```

## Common Hash Types

| Code | Algorithm | Example |
|------|-----------|---------|
| 0 | MD5 | 5f4dcc3b5aa765d61d8327deb882cf99 |
| 100 | SHA1 | 5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8 |
| 1400 | SHA256 | 5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8 |
| 1700 | SHA512 | b109f3bbbc244eb82441917ed06d618b9008dd09b3befd1b5e07394c706a8bb980b1d7785e5976ec049b46df5f1326af5a2ea6d103fd07c95385ffab0cacbc86 |
| 1000 | NTLM | 8846f7eaee8fb117ad06bdd830b7586c |

## Performance Tips

1. **Single hash?** → Use `GetHashByOriginalHash()` (O(1))
2. **Multiple hashes?** → Use `FindHashesByHashSum()` (single scan)
3. **Large datasets?** → Generators prevent OOM
4. **Don't need all results?** → Use `break` for early termination

## Key Format
```
krkn:{hashType}:{hexEncodedSHA256OfOriginalHash}
```

Example:
```
Original: 5f4dcc3b5aa765d61d8327deb882cf99
SHA256:   5f4dcc3b... → a1b2c3d4e5f6...
Key:      krkn:0:a1b2c3d4e5f6...
```

## Error Handling
```go
hash, err := db.GetHashByOriginalHash("...", 0)
if err != nil {
    if err == badger.ErrKeyNotFound {
        // Hash not in database
    } else {
        // Other error
    }
}
```

## Thread Safety
✅ All methods are thread-safe  
✅ Can be called from multiple goroutines  
✅ Mutex-protected internally

## Memory Usage
- **Direct lookup:** O(1) - Single hash
- **Batch search:** O(n) - HashMap of search hashes
- **Iteration:** O(1) - One hash at a time (generator)

## Examples Location
```bash
examples/basic_usage.go       # Basic operations
examples/performance_demo.go  # Performance comparison
```

## Common Patterns

### Batch Import
```go
for _, item := range importList {
    hash := kdb.NewHash(item.Hash, item.Value, item.Type)
    hash.Store()
}
```

### Crack Check
```go
hash, err := db.GetHashByOriginalHash(unknownHash, hashType)
if err == nil {
    fmt.Printf("Cracked! Password: %s\n", hash.Value)
}
```

### Export All
```go
for hash := range db.GetHashesByHashType(0) {
    fmt.Printf("%s:%s\n", hash.Hash, hash.Value)
}
```

### Batch Crack
```go
unknownHashes := []string{...}
for hash := range db.FindHashesByHashSum(unknownHashes, 0) {
    fmt.Printf("Found: %s = %s\n", hash.Hash, hash.Value)
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

