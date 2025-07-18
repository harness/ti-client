package utils

import (
	"strconv"

	"github.com/cespare/xxhash/v2"
	"github.com/harness/ti-client/types"
)

func ChainChecksum(chain []types.FilehashPair) uint64 {
	var checksum uint64 = 0
	for _, pair := range chain {
		candidate := []byte(pair.Path + ":" + strconv.FormatUint(pair.Checksum, 10))
		hash := xxhash.Sum64(candidate)
		checksum ^= hash
	}
	return checksum
}
