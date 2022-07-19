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
