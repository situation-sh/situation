package utils

import (
	"crypto/rand"
	"encoding/binary"
)

func fallBackRandomByte() byte {
	ch := make(chan byte)
	var b byte = 0
	go func() {
		for i := 0; i < 8; i++ {
			ch <- 0xFF
		}
	}()

	for i := 0; i < 8; i++ {
		select {
		case <-ch:
			// add 2^i
			b += 1 << i
		case <-ch:
			// nothing
		}
	}
	return b
}

func fallbackFillRandom(buffer []byte) {
	for i := 0; i < len(buffer); i++ {
		buffer[i] = fallBackRandomByte()
	}
}

func fillRandom(buffer []byte) {
	// Note that err == nil only if we read len(buffer) bytes.
	if _, err := rand.Read(buffer); err != nil {
		fallbackFillRandom(buffer)
	}
}

// RandUint16 returns an integer between 0 and max
func RandUint16(max uint16) uint16 {
	buffer := make([]byte, 2)
	fillRandom(buffer)
	return binary.BigEndian.Uint16(buffer) % max
}

func RandBytes(size int) []byte {
	buffer := make([]byte, size)
	fillRandom(buffer)
	return buffer
}

// RandomTCPPort returns a TCP port between a and b
// a <= port < b
func RandomTCPPort(a, b uint16) uint16 {
	if a > b {
		a, b = b, a
	}
	// e := uint16(rand.Intn(int(b - a)))
	e := RandUint16(b - a)
	return a + e
}
