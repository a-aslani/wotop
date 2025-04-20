package wotop_util

import (
	"regexp"
	"strings"
)

var matchFirstCapSnakeCase = regexp.MustCompile("(.)([A-Z][a-z]+)") // Regular expression to match the first capital letter in camel case.
var matchAllCapSnakeCase = regexp.MustCompile("([a-z\\d])([A-Z])")  // Regular expression to match all capital letters in camel case.

// SnakeCase converts a camel case string to snake case.
//
// This function uses regular expressions to identify camel case patterns
// and replaces them with underscores to create a snake case string.
//
// Parameters:
//   - str: The input string in camel case format.
//
// Returns:
//   - A string converted to snake case.
func SnakeCase(str string) string {
	snake := matchFirstCapSnakeCase.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCapSnakeCase.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// NormalizeDomainName normalizes a domain name by replacing dots with underscores
// and trimming any leading or trailing whitespace.
//
// Parameters:
//   - domain: The domain name to be normalized.
//
// Returns:
//   - A string representing the normalized domain name.
func NormalizeDomainName(domain string) string {
	return strings.TrimSpace(strings.Replace(domain, ".", "_", -1))
}

// ContainsStr checks if a slice of strings contains a specific string.
//
// This function iterates over the slice and returns true if the specified
// string is found.
//
// Parameters:
//   - s: The slice of strings to search.
//   - e: The string to look for.
//
// Returns:
//   - A boolean value indicating whether the string is found in the slice.
func ContainsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// ContainsInt checks if a slice of integers contains a specific integer.
//
// This function iterates over the slice and returns true if the specified
// integer is found.
//
// Parameters:
//   - s: The slice of integers to search.
//   - e: The integer to look for.
//
// Returns:
//   - A boolean value indicating whether the integer is found in the slice.
func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
