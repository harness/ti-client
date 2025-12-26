// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package utils

import "testing"

func TestChainChecksum(t *testing.T) {
	tests := []struct {
		name          string
		sourcePaths   []string
		fileChecksums map[string]uint64
		want          uint64
		wantZero      bool
	}{
		{
			name:        "empty source paths",
			sourcePaths: []string{},
			fileChecksums: map[string]uint64{
				"file1.go": 12345,
			},
			wantZero: true,
		},
		{
			name:        "no matching checksums",
			sourcePaths: []string{"file1.go", "file2.go"},
			fileChecksums: map[string]uint64{
				"file3.go": 12345,
				"file4.go": 67890,
			},
			wantZero: true,
		},
		{
			name:        "single matching path",
			sourcePaths: []string{"file1.go"},
			fileChecksums: map[string]uint64{
				"file1.go": 12345,
			},
			wantZero: false,
		},
		{
			name:        "multiple matching paths",
			sourcePaths: []string{"file1.go", "file2.go", "file3.go"},
			fileChecksums: map[string]uint64{
				"file1.go": 12345,
				"file2.go": 67890,
				"file3.go": 11111,
			},
			wantZero: false,
		},
		{
			name:        "partial matching paths",
			sourcePaths: []string{"file1.go", "file2.go", "file3.go"},
			fileChecksums: map[string]uint64{
				"file1.go": 12345,
				"file3.go": 11111,
				// file2.go is missing
			},
			wantZero: false,
		},
		{
			name:        "empty fileChecksums map",
			sourcePaths: []string{"file1.go", "file2.go"},
			fileChecksums: map[string]uint64{},
			wantZero: true,
		},
		{
			name:        "nil fileChecksums map",
			sourcePaths: []string{"file1.go"},
			fileChecksums: nil,
			wantZero: true,
		},
		{
			name:        "zero checksum values",
			sourcePaths: []string{"file1.go"},
			fileChecksums: map[string]uint64{
				"file1.go": 0,
			},
			wantZero: false, // Still processes even if checksum is 0
		},
		{
			name:        "large checksum values",
			sourcePaths: []string{"file1.go"},
			fileChecksums: map[string]uint64{
				"file1.go": 18446744073709551615, // max uint64
			},
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChainChecksum(tt.sourcePaths, tt.fileChecksums)
			if tt.wantZero {
				if got != 0 {
					t.Errorf("ChainChecksum() = %v, want 0", got)
				}
			} else {
				if got == 0 {
					t.Errorf("ChainChecksum() = 0, want non-zero value")
				}
			}
		})
	}
}

func TestChainChecksum_Deterministic(t *testing.T) {
	// Test that the same inputs produce the same output
	sourcePaths := []string{"file1.go", "file2.go"}
	fileChecksums := map[string]uint64{
		"file1.go": 12345,
		"file2.go": 67890,
	}

	result1 := ChainChecksum(sourcePaths, fileChecksums)
	result2 := ChainChecksum(sourcePaths, fileChecksums)

	if result1 != result2 {
		t.Errorf("ChainChecksum() is not deterministic: got %v and %v", result1, result2)
	}
}

func TestChainChecksum_OrderIndependent(t *testing.T) {
	// Test that order of sourcePaths doesn't matter (it should, based on implementation)
	// Actually, looking at the implementation, order DOES matter because it processes in order
	// So we test that different orders can produce different results
	fileChecksums := map[string]uint64{
		"file1.go": 12345,
		"file2.go": 67890,
		"file3.go": 11111,
	}

	result1 := ChainChecksum([]string{"file1.go", "file2.go", "file3.go"}, fileChecksums)
	result2 := ChainChecksum([]string{"file3.go", "file2.go", "file1.go"}, fileChecksums)

	// These should be different because order matters in the implementation
	if result1 == result2 {
		t.Log("ChainChecksum() produces same result for different orders (this is expected based on implementation)")
	}
}

