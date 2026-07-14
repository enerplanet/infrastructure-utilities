// Package utils provides shared utility functions for the common infrastructure layer.
package utils

import (
	"regexp"
	"strings"
)

// PasswordValidationError represents a single validation failure for a password field.
type PasswordValidationError struct {
	Field   string
	Message string
}

// weakPasswordRegex matches passwords that meet minimum strength requirements:
// at least 8 characters, at least one letter, and at least one digit or special character.
var weakPasswordRegex = regexp.MustCompile(`(?=.{8,})(?=.*[a-zA-Z])(?=.*[\d\W])`)

// ValidatePassword checks whether the given password meets minimum strength requirements.
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

	if !weakPasswordRegex.MatchString(password) {
		errs = append(errs, PasswordValidationError{
			Field:   "password",
			Message: "password does not meet minimum strength requirements (min 8 characters, at least one letter and one number or special character)",
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

// IsPasswordWeak checks a password for common weakness patterns: repeated characters
// (three or more in a row) and sequential character runs (e.g. "abc", "123").
// It returns true along with a reason if the password is considered weak.
func IsPasswordWeak(password string) (bool, string) {
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true, "password contains too many repeated characters"
		}
	}

	sequential := []string{
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij",
		"ijk", "jkl", "klm", "lmn", "mno", "nop", "opq", "pqr",
		"qrs", "rst", "stu", "tuv", "uvw", "vwx", "wxy", "xyz",
		"123", "234", "345", "456", "567", "678", "789",
	}
	for _, seq := range sequential {
		if strings.Contains(password, seq) {
			return true, "password contains sequential characters"
		}
	}

	return false, ""
}
