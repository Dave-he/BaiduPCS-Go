package netdisksign

import (
	"BaiduPCS-Go/pcsutil/cachepool"
	"BaiduPCS-Go/pcsutil/converter"
	"bytes"
	"crypto/md5"
	"encoding/hex"
)

func DevUID(feature string) string {
	m := md5.New()
	m.Write(converter.ToBytes(feature))
	res := m.Sum(nil)
	resHex := cachepool.RawMallocByteSlice(34)
	hex.Encode(resHex[:32], res)
	resHex[32] = '|'
	resHex[33] = '0'
	return converter.ToString(bytes.ToUpper(resHex))
}
