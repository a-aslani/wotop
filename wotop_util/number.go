package wotop_util

import "math"

// Round rounds a floating-point number to the nearest integer.
//
// This function uses `math.Copysign` to handle rounding for both positive
// and negative numbers.
//
// Parameters:
//   - num: The floating-point number to be rounded.
//
// Returns:
//   - An integer representing the rounded value.
func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// ToFixed rounds a floating-point number to a specified number of decimal places.
//
// This function multiplies the input number by 10 raised to the power of the
// specified precision, rounds it to the nearest integer, and then divides it
// back to achieve the desired precision.
//
// Parameters:
//   - num: The floating-point number to be rounded.
//   - precision: The number of decimal places to round to.
//
// Returns:
//   - A floating-point number rounded to the specified precision.
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}
