package ja4

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	EXT_SERVER_NAME          = uint16(0x0000)
	EXT_SUPPORTED_VERSIONS   = uint16(0x002b)
	EXT_SIGNATURE_ALGORITHMS = uint16(0x000d)
	EXT_ALPN                 = uint16(0x0010)
)

// extensionsToString converts multiple slices of uint16 extensions into a single string to match
// the format requested by the JA4+ suite
// If you give as input [0x0001, 0x0002] and [0x0003], it will return "0001,0002_0003"
func extensionsToString(exts ...[]uint16) string {
	concat := make([]string, 0)
	for _, ext := range exts {
		extHex := make([]string, 0)
		for _, e := range ext {
			extHex = append(extHex, fmt.Sprintf("%04x", e))
		}
		concat = append(concat, strings.Join(extHex, ","))
	}
	hash := sha256.Sum256([]byte(strings.Join(concat, "_")))
	return fmt.Sprintf("%x", hash[:6])
}

// isGREASE checks whether the given 16-bit TLS ID (cipher suite, extension, etc.)
// is a GREASE value as defined in RFC 8701.
func isGREASE(id uint16) bool {
	return id&0x0F0F == 0x0A0A
}

// alpn returns the ALPN protocol name as a two-character hex string.
// If the input is empty or too short, it returns "00".
// If the input is longer than 2 bytes, it returns the first and last byte if they are printable,
// otherwise bytes are converted to hex and then we take the first and last hex character.
func alpn(data []byte) string {
	if len(data) < 2 {
		return "00"
	}
	var out strings.Builder
	for i, b := range []byte{data[0], data[len(data)-1]} {
		if b < 127 {
			out.WriteByte(b)
		} else {
			h := hex.EncodeToString([]byte{b})
			out.WriteByte(h[i])
		}
	}
	return out.String()
}

func parseExtension(raw []byte) (uint16, []byte, error) {
	size := len(raw)
	if size < 4 {
		return 0, raw, fmt.Errorf("invalid extension length: %d", len(raw))
	}
	extType := binary.BigEndian.Uint16(raw[:2])
	extLen := binary.BigEndian.Uint16(raw[2:4])
	if size < int(extLen+4) {
		return 0, raw, fmt.Errorf("extension length %d exceeds raw data length %d", extLen, size)
	}
	return extType, raw[4 : 4+extLen], nil
}

func parseALPN(raw []byte) (string, error) {
	size := len(raw)
	//  ┌────────────────────────────────────────────┐
	//  │ Protocol Name List Length (2 bytes)        │ ← Length of list that follows
	//  ├────────────────────────────────────────────┤
	//  │ Protocol Name Length (1 byte)              │
	//  ├────────────────────────────────────────────┤
	//  │ Protocol Name (variable)                   │ ← e.g., "http/1.1" or "h2"
	//  ├────────────────────────────────────────────┤
	//  │ [Repeat for each protocol name]            │
	//  └────────────────────────────────────────────┘
	if size < 2 {
		return "", fmt.Errorf("ALPN data too short: %d bytes", len(raw))
	}
	protoCount := binary.BigEndian.Uint16(raw[:2])
	if protoCount == 0 {
		return "00", nil // No protocols
	}
	if size < 3 {
		return "", fmt.Errorf("ALPN protocol name length byte missing")
	}
	protoNameLen := int(raw[2])
	if protoNameLen == 0 {
		return "00", nil // No protocol name
	}
	if protoNameLen+3 > size {
		return "00", fmt.Errorf("ALPN protocol name length %d exceeds raw data length %d", protoNameLen, size)
	}
	return alpn(raw[3 : 3+protoNameLen]), nil
}

func supportTLS13(raw []byte) bool {
	// ┌────────────────────────────────────────────┐
	// │ Supported Versions Length (1 byte)         │ ← Length of list in bytes (must be even, not in ServerHello)
	// ├────────────────────────────────────────────┤
	// │ Supported Version 1 (2 bytes)              │ ← e.g. 0x0303 = TLS 1.2
	// ├────────────────────────────────────────────┤
	// │ Supported Version 2 (2 bytes)              │
	// ├────────────────────────────────────────────┤
	// │ ...                                        │ ← more versions
	// └────────────────────────────────────────────┘
	size := len(raw)
	if size < 2 {
		return false
	}
	if size == 2 {
		// in case of ServerHello, the extension contains a single tuple which is
		// the version of TLS
		if raw[0] == 0x03 && raw[1] == 0x04 {
			return true // TLS 1.3
		}
		return false // Not TLS 1.3
	}
	versionCount := int(raw[0])
	if versionCount == 0 {
		return false
	}
	for i := 1; i < size-1; i += 2 {
		if raw[i] == 0x03 && raw[i+1] == 0x04 {
			return true
		}
	}
	return false
}

func parseSignatureAlgorithms(raw []byte) ([]uint16, error) {
	//  ┌────────────────────────────────────────────┐
	//  │ Signature Algorithms Length (2 bytes)      │ ← Length of list in bytes (must be even)
	//  ├────────────────────────────────────────────┤
	//  │ Signature Algorithm 1 (2 bytes)            │ ← e.g. 0x0403 = rsa_pkcs1_sha256
	//  ├────────────────────────────────────────────┤
	//  │ Signature Algorithm 2 (2 bytes)            │ ← e.g. 0x0503 = rsa_pkcs1_sha384
	//  ├────────────────────────────────────────────┤
	//  │ ...                                        │ ← more pairs
	//  └────────────────────────────────────────────┘
	size := len(raw)
	sigAlgs := make([]uint16, 0)
	if size < 2 {
		return nil, fmt.Errorf("signature algorithms data too short: %d bytes", size)
	}
	sigAlgLen := binary.BigEndian.Uint16(raw[:2])
	if int(sigAlgLen) > size-2 {
		return nil, fmt.Errorf("signature algorithms length %d exceeds raw data length %d", sigAlgLen, size-2)
	}

	sigAlgCount := sigAlgLen / 2
	sigAlgs = make([]uint16, sigAlgCount)
	for j := uint16(0); j < sigAlgCount; j++ {
		sigAlgs[j] = binary.BigEndian.Uint16(raw[2*(j+1) : 2*(j+2)])
	}
	return sigAlgs, nil
}
