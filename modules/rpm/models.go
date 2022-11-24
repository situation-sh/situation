//go:build linux
// +build linux

package rpm

import (
	"encoding/binary"
	"time"
	"unicode/utf8"
)

type Pkg struct {
	Hnum uint   `db:"hnum"`
	Blob []byte `db:"blob"`
}

type Install struct {
	Key  []byte `db:"key"`
	Hnum uint   `db:"hnum"`
	Idx  uint   `db:"idx"`
}

func (a *Install) Parse() time.Time {
	ts := int64(binary.LittleEndian.Uint32(a.Key[:4]))
	return time.Unix(ts, 0)
}

func (p *Pkg) Value(storeOffset uint32, typ uint32, off uint32, cnt uint32) interface{} {
	store := p.Blob[storeOffset:]
	switch typ {
	case RPM_NULL_TYPE:
		return nil
	case RPM_CHAR_TYPE:
		r, _ := utf8.DecodeRune(store[off : off+cnt])
		return r
	case RPM_INT8_TYPE:
		return uint8(store[off])
	case RPM_INT16_TYPE:
		return binary.BigEndian.Uint16(store[off : off+2*cnt])
	case RPM_INT32_TYPE:
		return binary.BigEndian.Uint32(store[off : off+4*cnt])
	case RPM_INT64_TYPE:
		return binary.BigEndian.Uint64(store[off : off+8*cnt])
	case RPM_STRING_TYPE:
		n := uint32(len(store))
		i := off
		for ; (i < n) && (store[i] != 0); i++ {
		}
		return string(store[off : i+1])
	case RPM_BIN_TYPE:
		return store[off : off+cnt]
	default:
		// we currently do not handle other types

	}
	return nil
}

func (p *Pkg) Parse() map[string]interface{} {
	out := make(map[string]interface{})
	nIndex := binary.BigEndian.Uint32(p.Blob[:4])
	// hSize := binary.BigEndian.Uint32(data[4:8])
	headerSize := uint32(16)
	storeOffset := 8 + headerSize*nIndex

	for i := uint32(0); i < nIndex; i++ {
		header := p.Blob[8+headerSize*i : 8+headerSize*(i+1)]

		tag := binary.BigEndian.Uint32(header[:4])
		typ := binary.BigEndian.Uint32(header[4:8])
		off := binary.BigEndian.Uint32(header[8:12])
		cnt := binary.BigEndian.Uint32(header[12:16])

		switch tag {
		case RPMTAG_NAME:
			out["name"] = p.Value(storeOffset, typ, off, cnt)
		case RPMTAG_VERSION:
			out["version"] = p.Value(storeOffset, typ, off, cnt)
		case RPMTAG_RELEASE:
			out["release"] = p.Value(storeOffset, typ, off, cnt)
		case RPMTAG_VENDOR:
			out["vendor"] = p.Value(storeOffset, typ, off, cnt)
		default:
		}
	}
	return out
}
