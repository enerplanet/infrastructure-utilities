# Bloom Filter — Compromised Password Check

This directory contains [`bloomfilter.py`](bloomfilter.py), a tool that builds a **Bloom filter** from a known-compromised password list and exports it as a compact binary file. The binary is consumed at runtime by the Go backend (via [`//go:embed`](../common/pkg/utils/password.go:12)) to reject compromised passwords during registration.

## What is a Bloom Filter?

A [Bloom filter](https://en.wikipedia.org/wiki/Bloom_filter) is a space-efficient probabilistic data structure that answers _"is this item in the set?"_ with:

- **No false negatives** — if the filter says an item is _not_ present, it definitely isn't.
- **Configurable false positives** — the filter may say an item _is_ present when it isn't, but the rate is bounded by a target (`fp_rate`, default 1 %).

It works by mapping each item to `k` bit positions in a fixed-size bit array using `k` hash functions. To check membership, you verify that **all** `k` bits are set.

## How This Implementation Works

### 1. Loading & Filtering

[`UnifiedBloomGenerator.process()`](bloomfilter.py:55) loads passwords from either a local file or a URL (default: [SecLists xato-net-10-million-passwords](https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/xato-net-10-million-passwords-1000000.txt)). It:

- Strips whitespace and deduplicates entries.
- Filters out passwords shorter than `min_length` (default: 10).
- Optionally merges **compound guesswords** — all pairwise permutations of a custom word list (e.g. `"EnerPlanet"` + `"123"` → `"EnerPlanet123"`, `"123EnerPlanet"`).

### 2. Optimal Parameter Calculation

Given the number of unique items `n` and the target false-positive rate `p`, the tool computes optimal Bloom filter parameters:

```
m = - (n * ln(p)) / (ln(2)²)    ← bit array size
k = (m / n) * ln(2)             ← number of hash functions
```

These formulas minimise the bit array size for the desired `p`. See [`bloomfilter.py`](bloomfilter.py:96-98).

### 3. Hashing — Kirsch-Mitzenmacher Double-Hashing

Rather than using `k` independent hash functions (expensive), the implementation uses the **Kirsch-Mitzenmacher** double-hashing trick:

```python
hash_a = first_4_bytes(SHA256(item))
hash_b = next_4_bytes(SHA256(item))

for i in range(k):
    bit_index = (hash_a + i * hash_b) % m
```

This produces `k` uniformly distributed indices from just two 32-bit hash values. The Go consumer in [`password.go`](../common/pkg/utils/password.go:87-92) uses the exact same algorithm.

### 4. Binary Export Format

The filter is written to a binary file with a 13-byte header followed by the raw bit array:

| Offset | Size | Field | Description                          |
| ------ | ---- | ----- | ------------------------------------ |
| 0      | 4 B  | `n`   | Number of items inserted (uint32 BE) |
| 4      | 8 B  | `m`   | Bit array size in bits (uint64 BE)   |
| 12     | 1 B  | `k`   | Number of hash functions (uint8)     |
| 13+    | —    | bits  | Bit array, `ceil(m/8)` bytes         |

The Go side parses this header in [`loadBloomFilter()`](../common/pkg/utils/password.go:43-68).

### 5. Self-Test

After building the filter, [`_self_test()`](bloomfilter.py:129) picks 50 random passwords from the input set and verifies they all pass the filter. This catches logic errors early.

## Usage

```bash
# Run with defaults (downloads SecLists, outputs to ../common/pkg/utils/password_filter.bin)
python bloomfilter.py

# Customise
python -c "
from bloomfilter import UnifiedBloomGenerator
g = UnifiedBloomGenerator(min_length=8, fp_rate=0.001)
g.process(
    source='my_passwords.txt',
    export_txt=True,
    txt_output='filtered_list.txt',
    bin_output='password_filter.bin',
    custom_guesswords=['company', 'product', '2026'],
)
"
```

## Pipeline Overview

```
┌──────────────────────┐     ┌──────────────────┐     ┌──────────────────────┐
│  Password breach     │────▶│  bloomfilter.py  │────▶│  password_filter.bin │
│  list (URL or file)  │     │  (Python)        │     │  (13 B header + bits)│
└──────────────────────┘     └──────────────────┘     └──────────────────────┘
                                                              │
                                                              │ //go:embed
                                                              ▼
                                                      ┌──────────────────────┐
                                                      │  password.go         │
                                                      │  ValidatePassword()  │
                                                      └──────────────────────┘
```

The binary is embedded into the Go binary at compile time via [`//go:embed password_filter.bin`](../common/pkg/utils/password.go:12), so no file I/O is needed at runtime.

## Dependencies

- **Python 3.8+** — no third-party packages required (uses only `hashlib`, `math`, `urllib.request` from the standard library).
- **Go** — the consumer uses only the standard library (`crypto/sha256`, `encoding/binary`, `sync`).
