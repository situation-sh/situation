//go:build windows

package utils

import (
	"github.com/winlabs/gowin32"
)

func splitBuffer(buffer string) []string {
	out := make([]string, 0)
	inQuotes := false
	tmpBuffer := make([]rune, 0)
	for _, c := range buffer {
		if c == 32 { // space
			if inQuotes {
				// append the space
				tmpBuffer = append(tmpBuffer, c)
			} else {
				// flush
				out = append(out, string(tmpBuffer))
				tmpBuffer = nil
			}
		} else if c == 34 { // quote
			// toggle state and ignore character
			inQuotes = !inQuotes
		} else {
			// append
			tmpBuffer = append(tmpBuffer, c)
		}

	}
	return out
}

func GetCmd(pid int) ([]string, error) {
	// buffer is a string
	buffer, err := gowin32.GetProcessCommandLine(uint(pid))
	if err != nil {
		return nil, err
	}
	slices := splitBuffer(buffer)
	// slices := strings.Split(buffer, " ")
	// specify max capacity to len(slices)
	out := make([]string, 0, len(slices))
	for _, s := range slices {
		if len(s) > 0 {
			out = append(out, s)
		}
	}
	return out, nil
}
