package ja4

import (
	"encoding/hex"
	"fmt"
)

//  ┌────────────────────────────────────────────┐
//  │ Server Version (2 bytes)                   │ ← TLS version (e.g. 0x0303 for 1.2)
//  ├────────────────────────────────────────────┤
//  │ Random (32 bytes)                          │ ← Random nonce
//  ├────────────────────────────────────────────┤
//  │ Session ID Length (1 byte)                 │
//  ├────────────────────────────────────────────┤
//  │ Session ID (variable)                      │
//  ├────────────────────────────────────────────┤
//  │ Cipher Suite (2 bytes)                     │ ← e.g. 0xC02F for TLS_ECDHE_RSA...
//  ├────────────────────────────────────────────┤
//  │ Compression Method (1 byte)                │ ← usually 0x00 (null)
//  ├────────────────────────────────────────────┤
//  │ Extensions Length (2 bytes)                │ ← total length of extensions
//  └────────────────────────────────────────────┘
//         |
//         ▼

// JA4S parses a TLS ServerHello and returns the JA4S fingerprint
// We skip the TLS Record Layer
func JA4S(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty data")
	}
	offset := 0

	//  TLS Record Layer (5 bytes)
	//  ┌────────┬─────────────┬─────────────┐
	//  │ 0x16   │ Version     │ Length      │
	//  │ Handsh │ (2 bytes)   │ (2 bytes)   │
	//  └────────┴─────────────┴─────────────┘
	//        |      \__ e.g. 0x0303 (TLS 1.2), 0x0304 (TLS 1.3)
	//        \_ Record Type: 0x16 = Handshake
	if data[offset] == 0x16 {
		offset += 5
	}

	//  Handshake Layer (ServerHello starts here)
	//  ┌────────┬──────────────────────────────┐
	//  │ 0x02   │ Length (3 bytes)             │
	//  │ MsgTyp │ Total handshake message size │
	//  └────────┴──────────────────────────────┘

	if data[offset] != 0x02 {
		return "", fmt.Errorf("not a ServerHello (expected handshake type 0x02)")
	}
	// handshakeLen := int(data[offset+1])<<16 | int(data[offset+2])<<8 | int(data[offset+3])
	offset += 4

	//  ┌────────────────────────────────────────────┐
	//  │ Server Version (2 bytes)                   │ ← TLS version (e.g. 0x0303 for 1.2)
	//  ├────────────────────────────────────────────┤
	//  │ Random (32 bytes)                          │ ← Random nonce
	//  ├────────────────────────────────────────────┤
	//  │ Session ID Length (1 byte)                 │
	//  ├────────────────────────────────────────────┤
	//  │ Session ID (variable)                      │
	//  ├────────────────────────────────────────────┤
	//  │ Cipher Suite (2 bytes)                     │ ← e.g. 0xC02F for TLS_ECDHE_RSA...
	//  ├────────────────────────────────────────────┤
	//  │ Compression Method (1 byte)                │ ← usually 0x00 (null)
	//  ├────────────────────────────────────────────┤
	//  │ Extensions Length (2 bytes)                │ ← total length of extensions
	//  └────────────────────────────────────────────┘
	serverVersion := data[offset : offset+2]
	offset += 2

	// Skip random (32 bytes)
	offset += 32

	// Session ID
	sessLen := int(data[offset])
	offset += 1 + sessLen

	// Cipher Suite
	cipherSuite := data[offset : offset+2]
	cipherSuiteHex := hex.EncodeToString(cipherSuite)
	offset += 2

	// Compression method
	offset += 1

	// Check if extensions exist
	if offset >= len(data) {
		// No extensions
		return buildJA4S("t", serverVersion, cipherSuiteHex, []uint16{}, "00"), nil
	}

	extLen := int(data[offset])<<8 | int(data[offset+1])
	offset += 2

	extEnd := offset + extLen
	if extEnd > len(data) {
		return "", fmt.Errorf("extensions length exceeds data length")
	}
	// var extTypes []uint16
	extensions := make([]uint16, 0)
	// alpnByte := byte(0x00)
	protoName := "00"

	//  Extensions (variable)
	//  ┌──────────────┬──────────────┬─────────────────────────────┐
	//  │ Ext. Type    │ Ext. Length  │ Ext. Data                   │
	//  │ (2 bytes)    │ (2 bytes)    │ (variable)                  │
	//  ├──────────────┼──────────────┴─────────────────────────────┤
	//  │ Example: 0x0010 (ALPN) → may include h2 / http/1.1 etc.   │
	//  ├──────────────┼──────────────┬─────────────────────────────┤
	//  │ Other extensions in order (e.g., supported_groups, etc.)  │
	//  └───────────────────────────────────────────────────────────┘
	for offset+4 <= extEnd {
		extType, extData, err := parseExtension(data[offset:])
		if err != nil {
			return "", fmt.Errorf("error parsing extension: %w", err)
		}
		extDataLen := len(extData)

		extensions = append(extensions, extType)

		switch extType {
		case EXT_ALPN:
			protoName, err = parseALPN(extData)
			if err != nil {
				return "", fmt.Errorf("error parsing ALPN extension: %w", err)
			}
		case EXT_SUPPORTED_VERSIONS:
			if supportTLS13(extData) {
				serverVersion = []byte{0x03, 0x04} // TLS 1.3
			}
		}

		offset += 4 + extDataLen
	}

	return buildJA4S("t", serverVersion, cipherSuiteHex, extensions, protoName), nil
}

// buildJA4S assembles the fingerprint string
func buildJA4S(proto string, version []byte, cipher string, extensions []uint16, alpn string) string {
	// TLS version: 0x03 0x03 = TLS 1.2, 0x03 0x04 = TLS 1.3
	tlsVer := "12"
	if len(version) >= 2 && version[0] == 0x03 && version[1] == 0x04 {
		tlsVer = "13"
	}

	ja4s_a := fmt.Sprintf("%s%s%02x%s", proto, tlsVer, len(extensions), alpn)
	ja4s_c := extensionsToString(extensions)

	return fmt.Sprintf("%s_%s_%s", ja4s_a, cipher, ja4s_c)
}
