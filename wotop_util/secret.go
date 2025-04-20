package wotop_util

import (
	"math/rand"
	"time"
)

var letterRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // Characters used for generating random keys.
var numberRunes = "0123456789"                                                     // Characters used for generating random numbers.

// GenerateKey generates a random alphanumeric string of the specified length.
//
// This function uses a seeded random number generator to create a string
// consisting of characters from the `letterRunes` set.
//
// Parameters:
//   - n: The length of the key to be generated.
//
// Returns:
//   - A string containing the randomly generated key.
func GenerateKey(n int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[seededRand.Intn(len(letterRunes))]
	}
	return string(b)
}

// GenerateRandomNumber generates a random numeric string of the specified length.
//
// This function uses a seeded random number generator to create a string
// consisting of characters from the `numberRunes` set.
//
// Parameters:
//   - n: The length of the numeric string to be generated.
//
// Returns:
//   - A string containing the randomly generated numeric value.
func GenerateRandomNumber(n int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = numberRunes[seededRand.Intn(len(numberRunes))]
	}
	return string(b)
}
