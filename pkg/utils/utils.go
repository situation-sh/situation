// Package utils stores common utilities
package utils

import "slices"

import "sync"

func GetKeys(m map[string]any) []string {
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
	return slices.Contains(slice, s)
}

func MergeChannels[T any](channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	wg.Add(len(channels))
	for _, channel := range channels {
		ch := channel // capture range variable
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
