package wotop_util

import gonanoid "github.com/matoous/go-nanoid"

// alphabet defines the set of characters used to generate IDs.
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// GenerateID generates a unique ID of the specified length.
//
// This function uses the `gonanoid` library to create a random string
// consisting of characters from the predefined `alphabet`.
//
// Parameters:
//   - n: The length of the ID to be generated.
//
// Returns:
//   - A string representing the generated ID. If an error occurs during
//     generation, an empty string is returned.
func GenerateID(n int) string {
	ID, err := gonanoid.Generate(alphabet, n)
	if err != nil {
		return ""
	}

	return ID
}
