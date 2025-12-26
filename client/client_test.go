// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package client

import (
	"fmt"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "error with code and message",
			err:      &Error{Code: 404, Message: "Not Found"},
			expected: "404: Not Found",
		},
		{
			name:     "error with code only",
			err:      &Error{Code: 500, Message: ""},
			expected: "500: ",
		},
		{
			name:     "error with message only",
			err:      &Error{Code: 0, Message: "Something went wrong"},
			expected: "0: Something went wrong",
		},
		{
			name:     "empty error",
			err:      &Error{Code: 0, Message: ""},
			expected: "0: ",
		},
		{
			name:     "error with special characters",
			err:      &Error{Code: 400, Message: "Bad Request: Invalid input"},
			expected: "400: Bad Request: Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_ImplementsErrorInterface(t *testing.T) {
	var err error = &Error{Code: 500, Message: "Internal Server Error"}
	if err == nil {
		t.Error("Error should implement error interface")
	}

	expected := "500: Internal Server Error"
	if err.Error() != expected {
		t.Errorf("Error.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestError_CanBeUsedAsError(t *testing.T) {
	// Test that Error can be used in standard error handling patterns
	err := &Error{Code: 404, Message: "Not Found"}

	// Test error wrapping
	wrapped := fmt.Errorf("wrapped: %w", err)
	if wrapped == nil {
		t.Error("Error should be wrappable")
	}

	// Test error unwrapping
	unwrapped := fmt.Errorf("unwrapped: %w", err)
	if unwrapped.Error() == "" {
		t.Error("Error should be unwrappable")
	}
}

