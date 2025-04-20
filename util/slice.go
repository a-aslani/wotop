package util

import (
	"math/rand"
	"time"
)

// ToSliceAny converts a slice of any type to a slice of empty interface values.
//
// This function iterates over the input slice and appends each element
// to a new slice of type `[]any`.
//
// Type Parameters:
//   - T: The type of elements in the input slice.
//
// Parameters:
//   - objs: A slice of type `T` to be converted.
//
// Returns:
//   - A slice of type `[]any` containing the elements of the input slice.
func ToSliceAny[T any](objs []T) []any {
	datas := make([]any, 0)
	for _, obj := range objs {
		datas = append(datas, obj)
	}
	return datas
}

// GetRandomItem selects a random item from a slice.
//
// This function seeds the random number generator with the current time
// and generates a random index to pick an element from the input slice.
//
// Type Parameters:
//   - T: The type of elements in the input slice.
//
// Parameters:
//   - slice: A slice of type `T` from which a random item will be selected.
//
// Returns:
//   - An element of type `T` randomly selected from the input slice.
func GetRandomItem[T any](slice []T) T {
	rand.Seed(time.Now().UnixNano())     // seed or it will be set to 1
	randomIndex := rand.Intn(len(slice)) // generate a random int in the range 0 to len(slice)-1
	pick := slice[randomIndex]           // get the value from the slice
	return pick
}
