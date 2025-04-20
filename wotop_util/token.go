package wotop_util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

// GenerateToken generates a unique token based on the provided email string.
//
// This function first hashes the email using bcrypt with the default cost.
// If an error occurs during the hashing process, the program logs the error and exits.
// The resulting bcrypt hash is then further hashed using the MD5 algorithm,
// and the final token is returned as a hexadecimal-encoded string.
//
// Parameters:
//   - email: The email string to be used for generating the token.
//
// Returns:
//   - A string representing the generated token in hexadecimal format.
func GenerateToken(email string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(email), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Hash to store:", string(hash))

	hasher := md5.New()
	hasher.Write(hash)
	return hex.EncodeToString(hasher.Sum(nil))
}
