package wkutil

import "hash/crc32"

func HashCrc32(str string) uint32 {

	return crc32.ChecksumIEEE([]byte(str))
}
