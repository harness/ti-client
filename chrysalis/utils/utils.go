package utils

import (
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
)

func ChainChecksum(sourcePaths []string, fileChecksums map[string]uint64) uint64 {
	var candidates []string
	for _, path := range sourcePaths {
		if pathChecksum, exists := fileChecksums[path]; exists {
			candidates = append(candidates, path+":"+strconv.FormatUint(pathChecksum, 10))
		}
	}
	
	if len(candidates) == 0 {
		return 0
	}
	
	combined := strings.Join(candidates, "|")
	return xxhash.Sum64([]byte(combined))
}
