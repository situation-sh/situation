// Package utils stores common utilities
package utils

func GetKeys(m map[string]interface{}) []string {
	out := make([]string, len(m))
	k := 0
	for key := range m {
		out[k] = key
		k++
	}
	return out
}

// Includes check if a slice includes a given element (string)
func Includes[T comparable](slice []T, s T) bool {
	for _, x := range slice {
		if x == s {
			return true
		}
	}
	return false
}
