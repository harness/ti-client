package utils

import (
	"strconv"

	"github.com/cespare/xxhash/v2"
	"github.com/harness/ti-client/chrysalis/types"
)

func ChainChecksum(chain []types.FilehashPair) int64 {
	var checksum uint64 = 0
	for _, pair := range chain {
		candidate := []byte(pair.Path + ":" + strconv.FormatInt(pair.Checksum, 10))
		hash := xxhash.Sum64(candidate)
		checksum ^= hash
	}
	return int64(checksum)
}
