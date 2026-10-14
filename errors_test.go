package main

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInvalidArgs",
			err:      ErrInvalidArgs,
			expected: "invalid arguments",
		},
		{
			name:     "ErrInvalidOverride",
			err:      ErrInvalidOverride,
			expected: "invalid uv override",
		},
		{
			name:     "ErrTranslation",
			err:      ErrTranslation,
			expected: "command translation failed",
		},
		{
			name:     "ErrUVNotFound",
			err:      ErrUVNotFound,
			expected: "uv executable not found",
		},
		{
			name:     "ErrRecursiveExecution",
			err:      ErrRecursiveExecution,
			expected: "uv resolves to uvpip itself; refusing recursive execution",
		},
	}

	seen := make(map[string]bool)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("expected non-nil error for %s", tc.name)
			}
			if tc.err.Error() != tc.expected {
				t.Errorf("expected message %q, got %q", tc.expected, tc.err.Error())
			}
			if !errors.Is(tc.err, tc.err) {
				t.Errorf("expected errors.Is self-equivalence for %s", tc.name)
			}
			if seen[tc.err.Error()] {
				t.Errorf("duplicate error message detected across sentinel errors: %q", tc.err.Error())
			}
			seen[tc.err.Error()] = true
		})
	}
}
