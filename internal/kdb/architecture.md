# KrknDB Architecture

## Overview
KrknDB is a fast key-value lookup database optimized for storing and retrieving cryptographic hashes. It uses BadgerDB as the underlying storage engine with encryption support.

## Key Design

### Key Format
```
krkn:{hashType}:{hexEncodedSHA256Sum}
```

**Examples:**
- `krkn:0:5f4dcc3b5aa765d61d8327deb882cf99...` (MD5, type 0)
- `krkn:1400:5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8...` (SHA256, type 1400)
- `krkn:1700:b109f3bbbc244eb82441917ed06d618b9008dd09b3befd1b5e07394c706a8bb980b1d7785e5976ec049b46df5f1326af5a2ea6d103fd07c95385ffab0cacbc86...` (SHA512, type 1700)

### Why SHA256 of the Hash?
The original hash is hashed again with SHA256 to:
1. **Normalize key length**: All keys have consistent length regardless of original hash type
2. **Enable prefix searching**: Consistent format allows efficient prefix-based queries
3. **Indexing efficiency**: Fixed-length keys optimize BadgerDB's LSM tree structure

## Data Structure

### Hash Type
```go
type Hash struct {
    Hash     string // Original hash (e.g., MD5, SHA256, etc.)
    Sum      []byte // Hex-encoded SHA256 sum of the Hash (used in key)
    Value    string // The cracked password or secret
    HashType uint64 // Hash algorithm identifier (e.g., 0=MD5, 1400=SHA256)
    Key      []byte // Full key: krkn:{hashType}:{Sum}
}
```

## Query Methods & Performance

### 1. Store Hash
```go
hash := kdb.NewHash("5f4dcc3b5aa765d61d8327deb882cf99", "password", 0)
err := hash.Store()
```
- **Complexity:** O(log n)
- **Use case:** Adding new hash/password pairs

### 2. Direct Lookup (Single Hash)
```go
hash, err := db.GetHashByOriginalHash("5f4dcc3b5aa765d61d8327deb882cf99", 0)
```
- **Complexity:** O(1) - Direct key lookup
- **Use case:** Finding a single known hash
- **Best for:** 1-10 hashes

### 3. Batch Search (Multiple Hashes)
```go
hashes := []string{"hash1", "hash2", "hash3", ...}
for hash := range db.FindHashesByHashSum(hashes, 0) {
    // Process each found hash
}
```
- **Complexity:** O(m) where m = total hashes of this type
- **Strategy:** Single prefix scan + O(1) hashmap filtering
- **Use case:** Searching for multiple hashes at once
- **Best for:** 10+ hashes
- **Efficiency:** Scans the database ONCE, not once per hash

### 4. Type-based Iteration (All Hashes of Type)
```go
for hash := range db.GetHashesByHashType(0) {
    // Process each MD5 hash
}
```
- **Complexity:** O(m) where m = hashes of this type
- **Memory:** O(1) - Generator pattern, one hash at a time
- **Use case:** Processing all hashes of a specific algorithm
- **Features:** Supports early termination (break)

### 5. Prefix Search
```go
for hash := range db.SearchHashesByPrefix("5f4d", 0) {
    // Process matching hashes
}
```
- **Complexity:** O(k) where k = matching hashes
- **Use case:** Partial hash lookups, autocomplete
- **Note:** Searches the SHA256 sum prefix, not original hash

## Performance Characteristics

### When to Use Each Method

| Scenario | Method | Complexity | Reason |
|----------|--------|------------|--------|
| Find 1 hash | `GetHashByOriginalHash()` | O(1) | Direct lookup |
| Find 2-10 hashes | Either method works | O(1) to O(m) | Depends on dataset |
| Find 10+ hashes | `FindHashesByHashSum()` | O(m) | Single scan |
| Find all of type | `GetHashesByHashType()` | O(m) | Full iteration |
| Partial match | `SearchHashesByPrefix()` | O(k) | Prefix scan |

### Example Performance Comparison

**Scenario:** 1 million total hashes, 100k MD5 hashes (type 0), searching for 1000 specific hashes

**Option A: Individual Lookups**
```go
for _, h := range searchHashes {
    hash, _ := db.GetHashByOriginalHash(h, 0)
}
```
- Operations: 1000 × log(1M) ≈ 20,000 operations
- Database scans: 1000

**Option B: Batch Search (Implemented)**
```go
for hash := range db.FindHashesByHashSum(searchHashes, 0) {
    // Process
}
```
- Operations: 100k scans + 1000 O(1) checks ≈ 100k operations
- Database scans: 1 (single prefix scan)
- **Winner when:** Searching for 10+ hashes

## Generator Pattern (Go 1.23+ Iterators)

### Why Generators?
```go
for hash := range db.GetHashesByHashType(0) {
    if someCondition {
        break // Only fetched what we needed!
    }
}
```

**Benefits:**
1. **Memory efficient**: Only one hash in memory at a time
2. **Lazy evaluation**: Only fetches data as needed
3. **Early termination**: Can stop iteration anytime
4. **Clean syntax**: Standard Go `for range` loops
5. **Large datasets**: Can handle millions of hashes without OOM

## Thread Safety

All database operations are protected with mutex locks:
```go
kc.mu.Lock()
defer kc.mu.Unlock()
```

Safe for concurrent access from multiple goroutines.

## Storage Engine

- **Backend:** BadgerDB (LSM tree)
- **Encryption:** AES-256 (32-byte key required)
- **Compression:** Handled by BadgerDB
- **Transactions:** ACID compliant

## Use Cases

1. **Password cracking databases**: Store hash→password mappings
2. **Rainbow tables**: Efficient hash lookups
3. **Hash analysis**: Iterate through hash types
4. **Breach databases**: Search for compromised hashes
5. **Security research**: Large-scale hash storage and retrieval

