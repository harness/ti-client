// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package telemetryutils

import (
	"testing"

	"github.com/harness/ti-client/types"
)

func TestCountDistinctClasses(t *testing.T) {
	tests := []struct {
		name      string
		testCases  []*types.TestCase
		want       int
	}{
		{
			name: "empty slice",
			testCases: []*types.TestCase{},
			want: 0,
		},
		{
			name: "single test case",
			testCases: []*types.TestCase{
				{ClassName: "TestClass1"},
			},
			want: 1,
		},
		{
			name: "multiple test cases with unique classes",
			testCases: []*types.TestCase{
				{ClassName: "TestClass1"},
				{ClassName: "TestClass2"},
				{ClassName: "TestClass3"},
			},
			want: 3,
		},
		{
			name: "multiple test cases with duplicate classes",
			testCases: []*types.TestCase{
				{ClassName: "TestClass1"},
				{ClassName: "TestClass1"},
				{ClassName: "TestClass2"},
				{ClassName: "TestClass2"},
				{ClassName: "TestClass2"},
			},
			want: 2,
		},
		{
			name: "test cases with empty class names",
			testCases: []*types.TestCase{
				{ClassName: ""},
				{ClassName: "TestClass1"},
				{ClassName: ""},
			},
			want: 2, // empty string and TestClass1
		},
		{
			name: "all test cases with same class",
			testCases: []*types.TestCase{
				{ClassName: "TestClass1"},
				{ClassName: "TestClass1"},
				{ClassName: "TestClass1"},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountDistinctClasses(tt.testCases)
			if got != tt.want {
				t.Errorf("CountDistinctClasses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountDistinctSelectedClasses(t *testing.T) {
	tests := []struct {
		name  string
		tests []types.RunnableTest
		want  int
	}{
		{
			name:  "empty slice",
			tests: []types.RunnableTest{},
			want:  0,
		},
		{
			name: "single test",
			tests: []types.RunnableTest{
				{Class: "TestClass1"},
			},
			want: 1,
		},
		{
			name: "multiple tests with unique classes",
			tests: []types.RunnableTest{
				{Class: "TestClass1"},
				{Class: "TestClass2"},
				{Class: "TestClass3"},
			},
			want: 3,
		},
		{
			name: "multiple tests with duplicate classes",
			tests: []types.RunnableTest{
				{Class: "TestClass1"},
				{Class: "TestClass1"},
				{Class: "TestClass2"},
				{Class: "TestClass2"},
				{Class: "TestClass2"},
			},
			want: 2,
		},
		{
			name: "tests with empty class names",
			tests: []types.RunnableTest{
				{Class: ""},
				{Class: "TestClass1"},
				{Class: ""},
			},
			want: 2, // empty string and TestClass1
		},
		{
			name: "all tests with same class",
			tests: []types.RunnableTest{
				{Class: "TestClass1"},
				{Class: "TestClass1"},
				{Class: "TestClass1"},
			},
			want: 1,
		},
		{
			name: "tests with different fields but same class",
			tests: []types.RunnableTest{
				{Class: "TestClass1", Method: "test1"},
				{Class: "TestClass1", Method: "test2"},
				{Class: "TestClass1", Method: "test3"},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountDistinctSelectedClasses(tt.tests)
			if got != tt.want {
				t.Errorf("CountDistinctSelectedClasses() = %v, want %v", got, tt.want)
			}
		})
	}
}

