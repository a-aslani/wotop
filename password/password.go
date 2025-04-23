package password

import "golang.org/x/crypto/bcrypt"

//go:generate go run go.uber.org/mock/mockgen -destination hasher_mock.go -package password ./ Hasher

// Hasher defines an interface for password hashing and verification.
// Methods:
// - HashPassword: Hashes a plain text password and returns the hashed string.
// - CheckPasswordHash: Verifies if a plain text password matches a given hash.
type Hasher interface {
	// HashPassword hashes the given plain text password.
	// Parameters:
	// - password: The plain text password to hash.
	// Returns:
	// - string: The hashed password.
	// - error: An error if the hashing operation fails.
	HashPassword(password string) (string, error)

	// CheckPasswordHash verifies if the given plain text password matches the provided hash.
	// Parameters:
	// - password: The plain text password to verify.
	// - hash: The hashed password to compare against.
	// Returns:
	// - bool: True if the password matches the hash, false otherwise.
	CheckPasswordHash(password, hash string) bool
}

// BcryptHashing implements the Hasher interface using bcrypt for password hashing and verification.
type BcryptHashing struct {
	// Const defines the cost parameter for bcrypt hashing.
	// Higher values increase the computation time for hashing.
	Const int
}

// HashPassword hashes the given plain text password using bcrypt.
// Parameters:
// - password: The plain text password to hash.
// Returns:
// - string: The hashed password.
// - error: An error if the hashing operation fails.
func (b BcryptHashing) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.Const)
	return string(bytes), err
}

// CheckPasswordHash verifies if the given plain text password matches the provided bcrypt hash.
// Parameters:
// - password: The plain text password to verify.
// - hash: The bcrypt hashed password to compare against.
// Returns:
// - bool: True if the password matches the hash, false otherwise.
func (b BcryptHashing) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
