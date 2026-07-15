// Package utils provides shared utility functions for the common infrastructure layer.
package utils

import (
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"errors"
	"sync"
)

//go:embed password_filter.bin
var bloomFilterData []byte

const (
	// MinPasswordLength is the minimum allowed password length per NIST SP 800-63B (2026).
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum allowed password length per NIST SP 800-63B (2026).
	// NIST recommends supporting at least 64 characters.
	MaxPasswordLength = 128
)

// PasswordValidationError represents a single validation failure for a password field.
type PasswordValidationError struct {
	Field   string
	Message string
}

// bloomFilter holds the parsed bloom filter metadata and bit array.
type bloomFilter struct {
	mu       sync.RWMutex
	loaded   bool
	n        uint32 // number of items
	m        uint64 // bit array size
	k        uint8  // number of hash functions
	bitArray []byte
}

var globalBloom = &bloomFilter{}

// loadBloomFilter parses the embedded bloom filter binary once.
// Format: 4 bytes n (uint32 BE), 8 bytes m (uint64 BE), 1 byte k, then bit array.
func loadBloomFilter() error {
	globalBloom.mu.Lock()
	defer globalBloom.mu.Unlock()

	if globalBloom.loaded {
		return nil
	}

	data := bloomFilterData
	if len(data) < 13 {
		return errors.New("bloom filter data too short")
	}

	globalBloom.n = binary.BigEndian.Uint32(data[0:4])
	globalBloom.m = binary.BigEndian.Uint64(data[4:12])
	globalBloom.k = data[12]
	globalBloom.bitArray = data[13:]

	expectedBytes := (globalBloom.m + 7) / 8
	if uint64(len(globalBloom.bitArray)) < expectedBytes {
		return errors.New("bloom filter bit array truncated")
	}

	globalBloom.loaded = true
	return nil
}

// isCompromised checks whether the given password appears in the bloom filter
// of known compromised passwords. Returns true if the password is likely compromised.
func isCompromised(password string) bool {
	if err := loadBloomFilter(); err != nil {
		// If the bloom filter can't be loaded, err on the side of caution
		// and treat the password as not compromised (allow it through).
		return false
	}

	globalBloom.mu.RLock()
	m := globalBloom.m
	k := globalBloom.k
	bitArray := globalBloom.bitArray
	globalBloom.mu.RUnlock()

	// Generate hash indices using Kirsch-Mitzenmacher double-hashing
	// (same algorithm as the Python generator).
	digest := sha256.Sum256([]byte(password))
	hashA := binary.BigEndian.Uint32(digest[0:4])
	hashB := binary.BigEndian.Uint32(digest[4:8])

	for i := uint8(0); i < k; i++ {
		bitIndex := (uint64(hashA) + uint64(i)*uint64(hashB)) % m
		byteIndex := bitIndex / 8
		bitPosition := bitIndex % 8

		if byteIndex >= uint64(len(bitArray)) {
			return false
		}
		if bitArray[byteIndex]&(1<<bitPosition) == 0 {
			return false
		}
	}

	return true
}

// ValidatePassword checks whether the given password meets NIST SP 800-63B (2026)
// recommendations:
//   - Minimum 8 characters
//   - Maximum 128 characters (NIST recommends supporting at least 64)
//   - No composition rules (no required uppercase, lowercase, digits, or special chars)
//   - Checked against a bloom filter of known compromised passwords
//
// It returns a slice of validation errors; an empty slice means the password is valid.
func ValidatePassword(password string) []PasswordValidationError {
	var errs []PasswordValidationError

	if password == "" {
		errs = append(errs, PasswordValidationError{
			Field:   "password",
			Message: "password is required",
		})
		return errs
	}

	if len(password) < MinPasswordLength {
		errs = append(errs, PasswordValidationError{
			Field:   "password",
			Message: "password must be at least 8 characters",
		})
	}

	if len(password) > MaxPasswordLength {
		errs = append(errs, PasswordValidationError{
			Field:   "password",
			Message: "password must not exceed 128 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	if isCompromised(password) {
		errs = append(errs, PasswordValidationError{
			Field:   "password",
			Message: "this password has been compromised in a known data breach and cannot be used",
		})
	}

	return errs
}

// ValidatePasswordMatch compares password and confirmation, returning a validation
// error if they do not match, or nil if they are equal.
func ValidatePasswordMatch(password, confirmation string) *PasswordValidationError {
	if password != confirmation {
		return &PasswordValidationError{
			Field:   "password_confirmation",
			Message: "passwords do not match",
		}
	}
	return nil
}

// IsPasswordWeak checks a password against the bloom filter of known compromised
// passwords. Returns true along with a reason if the password appears in the
// compromised password list.
//
// This replaces the old sequential/repeated character checks with a NIST SP 800-63B
// (2026) compliant approach: the only weakness check is whether the password
// has been previously exposed in a known breach.
func IsPasswordWeak(password string) (bool, string) {
	if isCompromised(password) {
		return true, "this password has been compromised in a known data breach and cannot be used"
	}
	return false, ""
}
