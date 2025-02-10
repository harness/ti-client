package telemetryutils

import (
	"github.com/harness/ti-client/types"
)

func CountDistinctClasses(testCases []*types.TestCase) int {
	uniqueClasses := make(map[string]bool)

	for _, testCase := range testCases {
		uniqueClasses[testCase.ClassName] = true
	}

	return len(uniqueClasses)
}

func CountDistinctSelectedClasses(tests []types.RunnableTest) int {
	uniqueClasses := make(map[string]bool) // Map to track unique class names

	for _, test := range tests {
		uniqueClasses[test.Class] = true // Add class to map (duplicates will be ignored)
	}

	return len(uniqueClasses) // Return the count of unique keys in the map
}

func CountDistinctClasses(testCases []*types.TestCase) int {
	uniqueClasses := make(map[string]bool)

	for _, testCase := range testCases {
		uniqueClasses[testCase.ClassName] = true
	}

	return len(uniqueClasses)
}

func CountDistinctSelectedClasses(tests []types.RunnableTest) int {
	uniqueClasses := make(map[string]bool) // Map to track unique class names

	for _, test := range tests {
		uniqueClasses[test.Class] = true // Add class to map (duplicates will be ignored)
	}

	return len(uniqueClasses) // Return the count of unique keys in the map
}
