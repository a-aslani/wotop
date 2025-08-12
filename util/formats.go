package util

import (
	"fmt"
	"regexp"

	"github.com/nyaruka/phonenumbers"
)

// IsE164 checks if the phone number is already in E.164 format
func IsE164(number string) bool {
	// E.164 format: + followed by 8 to 15 digits
	re := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	return re.MatchString(number)
}

// NormalizePhone formats a phone number to E.164 if needed
func NormalizePhone(rawNumber, region string) (string, error) {
	// If already in E.164 format and valid, return as-is
	if IsE164(rawNumber) {
		num, err := phonenumbers.Parse(rawNumber, "")
		if err == nil && phonenumbers.IsValidNumber(num) {
			return rawNumber, nil
		}
	}

	// Parse and validate
	num, err := phonenumbers.Parse(rawNumber, region)
	if err != nil {
		return "", fmt.Errorf("failed to parse phone number: %v", err)
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number")
	}

	// Convert to E.164
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
