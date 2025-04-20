package wotop_util

import "encoding/json"

// MustJSON converts an object to its JSON string representation.
//
// This function marshals the given object into a JSON string. If an error occurs
// during the marshaling process, it is ignored, and the resulting string is returned.
//
// Parameters:
//   - obj: The object to be marshaled into JSON.
//
// Returns:
//   - A string containing the JSON representation of the object.
func MustJSON(obj any) string {
	bytes, _ := json.Marshal(obj)
	return string(bytes)
}
