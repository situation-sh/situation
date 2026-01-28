//go:build linux
// +build linux

package rpm

import (
	"bytes"
	"encoding/binary"
	"path"
	"unicode/utf8"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/utils"
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

func (a *Install) Parse() int64 {
	return int64(binary.LittleEndian.Uint32(a.Key[:4]))
}

func (p *Pkg) Value(storeOffset uint32, typ uint32, off uint32, cnt uint32) interface{} {
	var step uint32
	store := p.Blob[storeOffset:]
	switch typ {
	case RPM_NULL_TYPE:
		return nil
	case RPM_CHAR_TYPE:
		r, _ := utf8.DecodeRune(store[off : off+cnt])
		return r
	case RPM_INT8_TYPE:
		step = 1
		out := make([]uint8, cnt)
		for i := uint32(0); i < cnt; i++ {
			out[i] = uint8(store[off+i])
		}
		if cnt == 1 {
			return out[0]
		}
		return out
	case RPM_INT16_TYPE:
		step = 2
		out := make([]uint16, cnt)
		for i := uint32(0); i < cnt; i++ {
			out[i] = binary.BigEndian.Uint16(store[off+i*step : off+(i+1)*step])
		}
		if cnt == 1 {
			return out[0]
		}
		return out
	case RPM_INT32_TYPE:
		step = 4
		out := make([]uint32, cnt)
		for i := uint32(0); i < cnt; i++ {
			out[i] = binary.BigEndian.Uint32(store[off+i*step : off+(i+1)*step])
		}
		if cnt == 1 {
			return out[0]
		}
		return out
	case RPM_INT64_TYPE:
		step = 8
		out := make([]uint32, cnt)
		for i := uint32(0); i < cnt; i++ {
			out[i] = binary.BigEndian.Uint32(store[off+i*step : off+(i+1)*step])
		}
		if cnt == 1 {
			return out[0]
		}
		return out
	case RPM_STRING_TYPE:
		n := uint32(0)
		lens := int64(len(store))
		if lens <= int64(0xffffffff) {
			n = uint32(lens)
		}
		i := off
		for ; (i < n) && (store[i] != 0); i++ {
		}
		return string(store[off:i])
	case RPM_BIN_TYPE:
		return store[off : off+cnt]
	case RPM_STRING_ARRAY_TYPE:
		out := make([]string, cnt)
		end := int(off)
		start := 0

		for k := uint32(0); k < cnt; k++ {
			start = end
			end += bytes.IndexByte(store[end:], 0x0)
			out[k] = string(store[start:end])
			end += 1
		}
		return out
	default:
		// we currently do not handle other types

	}
	return nil
}

func reassembleFiles(basenames []string, dirnames []string, dirindexes []uint32) []string {
	out := make([]string, len(basenames))
	for i, f := range basenames {
		out[i] = path.Join(dirnames[dirindexes[i]], f)
	}
	return out
}

func (p *Pkg) Parse() *models.Package {
	pkg := models.Package{Manager: "rpm"}

	// out := make(map[string]interface{})
	nIndex := binary.BigEndian.Uint32(p.Blob[:4])
	// hSize := binary.BigEndian.Uint32(data[4:8])
	headerSize := uint32(16)
	storeOffset := 8 + headerSize*nIndex

	var basenames []string
	var dirnames []string
	var dirindexes []uint32

	for i := uint32(0); i < nIndex; i++ {
		header := p.Blob[8+headerSize*i : 8+headerSize*(i+1)]

		tag := binary.BigEndian.Uint32(header[:4])
		typ := binary.BigEndian.Uint32(header[4:8])
		off := binary.BigEndian.Uint32(header[8:12])
		cnt := binary.BigEndian.Uint32(header[12:16])

		switch tag {
		case RPMTAG_NAME:
			v := p.Value(storeOffset, typ, off, cnt)
			pkg.Name, _ = v.(string)
		case RPMTAG_VERSION:
			v := p.Value(storeOffset, typ, off, cnt)
			pkg.Version, _ = v.(string)
		case RPMTAG_VENDOR:
			v := p.Value(storeOffset, typ, off, cnt)
			pkg.Vendor, _ = v.(string)
		case RPMTAG_BASENAMES:
			switch v := p.Value(storeOffset, typ, off, cnt); t := v.(type) {
			case []string:
				basenames = t
			default:
			}
			// out["basenames"] = p.Value(storeOffset, typ, off, cnt)
		case RPMTAG_DIRNAMES:
			switch v := p.Value(storeOffset, typ, off, cnt); t := v.(type) {
			case []string:
				dirnames = t
			default:
			}
			// out["dirnames"] = p.Value(storeOffset, typ, off, cnt)
		case RPMTAG_DIRINDEXES:
			switch v := p.Value(storeOffset, typ, off, cnt); t := v.(type) {
			case []uint32:
				dirindexes = t
			default:
			}
		default:
		}
	}

	if len(basenames) > 0 && len(dirnames) > 0 && len(dirindexes) == len(basenames) {
		pkg.Files = utils.KeepLeaves(reassembleFiles(basenames, dirnames, dirindexes))
	}

	return &pkg
}
