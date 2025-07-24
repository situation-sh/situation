package ja4

import (
	"encoding/binary"
	"fmt"
	"slices"
)

func JA4(data []byte) (string, error) {
	// var err error
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

	//  ┌────────────────────────────────────────────┐
	//  │ Handshake Type (1 byte)                    │ ← 0x01 = ClientHello
	//  ├────────────────────────────────────────────┤
	//  │ Length (3 bytes)                           │ ← length of ClientHello body
	//  └────────────────────────────────────────────┘
	if data[offset] != 0x01 {
		return "", fmt.Errorf("not a ClientHello (expected handshake type 0x01)")
	}
	offset += 1

	// length := binary.BigEndian.Uint32(append([]byte{0x0}, data[offset:offset+3]...))
	offset += 3

	//  ┌────────────────────────────────────────────┐
	//  │ Client Version (2 bytes)                   │ ← e.g. 0x0303 for TLS 1.2
	//  ├────────────────────────────────────────────┤
	//  │ Random (32 bytes)                          │ ← Client random nonce
	//  ├────────────────────────────────────────────┤
	//  │ Session ID Length (1 byte)                 │
	//  ├────────────────────────────────────────────┤
	//  │ Session ID (variable)                      │ ← May be 0 length
	//  ├────────────────────────────────────────────┤
	//  │ Cipher Suites Length (2 bytes)             │ ← Length in bytes (must be even)
	//  ├────────────────────────────────────────────┤
	//  │ Cipher Suites (variable, 2 bytes each)     │ ← e.g. 0xC02F = TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	//  ├────────────────────────────────────────────┤
	//  │ Compression Methods Length (1 byte)        │ ← Usually 1
	//  ├────────────────────────────────────────────┤
	//  │ Compression Methods (1 byte each)          │ ← Usually 0x00 (null)
	//  ├────────────────────────────────────────────┤
	//  │ Extensions Length (2 bytes)                │ ← Total length of extensions
	//  ├────────────────────────────────────────────┤
	//  │ Extensions (variable)                      │ ← Each extension has: type + length + data
	//  └────────────────────────────────────────────┘
	clientVersion := data[offset : offset+2]
	offset += 2

	// Skip random (32 bytes)
	offset += 32

	// Skip session ID length and session ID
	sessionIDLen := data[offset]
	offset += 1 + int(sessionIDLen)

	// Cipher Suites
	cipherSuiteLen := binary.BigEndian.Uint16(data[offset : offset+2])
	// fmt.Printf("Cipher Suites Length: %d (%v)\n", cipherSuiteLen, data[offset:offset+2])
	offset += 2

	cipherSuiteCount := cipherSuiteLen / 2
	// fmt.Println("Cipher Suite Count:", cipherSuiteCount)
	cipherSuites := make([]uint16, 0, cipherSuiteCount)
	for i := 0; i < int(cipherSuiteCount); i++ {
		suite := binary.BigEndian.Uint16(data[offset : offset+2])
		if !isGREASE(suite) {
			cipherSuites = append(cipherSuites, suite)
		}
		// fmt.Printf("Cipher Suite: %d (%v)\n", cipherSuites[i], data[offset:offset+2])
		offset += 2
	}

	// Skip Compression Methods
	compressionMethodsLen := data[offset]
	offset += 1 + int(compressionMethodsLen)

	// extensions
	extensionsLen := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	extensionsEnd := offset + int(extensionsLen)
	if extensionsEnd > len(data) {
		return "", fmt.Errorf("extensions length exceeds data length")
	}
	extensions := make([]uint16, 0)

	sni := false
	protoName := "00"
	sigAlgs := make([]uint16, 0)
	var err error
	//  ┌────────────────────────────────────────────┐
	//  │ Extension Type (2 bytes)                   │ ← e.g. 0x000d = signature_algorithms
	//  ├────────────────────────────────────────────┤
	//  │ Extension Length (2 bytes)                 │ ← Length of Extension Data
	//  ├────────────────────────────────────────────┤
	//  │ Extension Data (variable)                  │ ← Depends on type
	//  └────────────────────────────────────────────┘
	for offset+4 <= extensionsEnd {
		extType := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2
		if !isGREASE(extType) {
			extensions = append(extensions, extType)
		}

		extLen := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		// use a local offset to parse extension
		k := offset
		switch extType {
		case EXT_SERVER_NAME: // server_name
			sni = true
		case EXT_SUPPORTED_VERSIONS: // supported_versions
			if supportTLS13(data[k:]) {
				clientVersion = []byte{0x03, 0x04} // TLS 1.3
			}
		case EXT_SIGNATURE_ALGORITHMS: // signature_algorithms
			sigAlgs, err = parseSignatureAlgorithms(data[k:])
			if err != nil {
				return "", fmt.Errorf("error parsing signature algorithms: %w", err)
			}
		case EXT_ALPN: // application_layer_protocol_negotiation
			protoName, err = parseALPN(data[k:])

		}
		// jump to next extension
		offset += int(extLen)
	}

	return buildJA4("t", clientVersion, sni, cipherSuites, extensions, sigAlgs, protoName), nil
}

// t13d1516h2_002f,0035,009c,009d,1301,1302,1303,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0012,0015,0017,001b,0023,002b,002d,0033,4469,ff01_0403,0804,0401,0503,0805,0501,0806,0601

func buildJA4(proto string, version []byte, sni bool, ciphers []uint16, extensions []uint16, signatures []uint16, alpn string) string {

	// TLS version: 0x03 0x03 = TLS 1.2, 0x03 0x04 = TLS 1.3
	tlsVer := "12"
	if len(version) >= 2 && version[0] == 0x03 && version[1] == 0x04 {
		tlsVer = "13"
	}

	// fmt.Printf("ALPN: %v\n", alpn)
	// if len(alpn) >= 2 {
	// 	alpn = string([]byte{alpn[0], alpn[len(alpn)-1]})
	// }

	s := "i"
	if sni {
		s = "d"
	}
	ja4a := fmt.Sprintf("%s%s%s%02d%02d%s", proto, tlsVer, s, len(ciphers), len(extensions), alpn)

	slices.Sort(ciphers)
	ja4b := extensionsToString(ciphers)

	slices.Sort(extensions)
	cleanedExtensions := make([]uint16, 0)
	for _, ext := range extensions {
		if ext != 0x0000 && ext != 0x0010 {
			cleanedExtensions = append(cleanedExtensions, ext)
		}
	}
	// fmt.Println("EXTENSIONS:", extensions)
	// fmt.Println("SIGNATURES:", signatures)
	// extsign := append(extensions, signatures...)
	ja4c := extensionsToString(cleanedExtensions, signatures)

	return fmt.Sprintf("%s_%s_%s", ja4a, ja4b, ja4c)
}
