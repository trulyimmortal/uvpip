package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateInputsSchemaBoundary enforces table-driven schema verification across
// all fields of CLIInputs (Args and UVOverride), asserting exact boundary limits and typed errors.
func TestValidateInputsSchemaBoundary(t *testing.T) {
	tempDir := t.TempDir()
	validAbsPath := filepath.Join(tempDir, "uv")
	if filepath.VolumeName(validAbsPath) == "" && !filepath.IsAbs(validAbsPath) {
		validAbsPath = "/" + validAbsPath
	}

	tests := []struct {
		name       string
		inputs     CLIInputs
		wantPass   bool
		wantTyped  error
		wantErrSub string
	}{
		// Args schema boundaries
		{
			name:       "nil args rejected",
			inputs:     CLIInputs{Args: nil},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: "pip command is empty",
		},
		{
			name:       "empty slice args rejected",
			inputs:     CLIInputs{Args: []string{}},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: "pip command is empty",
		},
		{
			name:       "blank command arg rejected",
			inputs:     CLIInputs{Args: []string{"   "}},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: "pip command is empty",
		},
		{
			name:     "single valid command arg accepted",
			inputs:   CLIInputs{Args: []string{"install"}},
			wantPass: true,
		},
		{
			name:     "typical command with flags accepted",
			inputs:   CLIInputs{Args: []string{"install", "-r", "requirements.txt", "--upgrade"}},
			wantPass: true,
		},
		{
			name:     "maximum boundary arg count accepted (512 args)",
			inputs:   CLIInputs{Args: makeArgList("pkg", maxArgCount)},
			wantPass: true,
		},
		{
			name:       "exceeded arg count rejected (513 args)",
			inputs:     CLIInputs{Args: makeArgList("pkg", maxArgCount+1)},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: fmt.Sprintf("argument count %d exceeds maximum limit (%d)", maxArgCount+1, maxArgCount),
		},
		{
			name:     "maximum boundary arg length accepted (4096 chars)",
			inputs:   CLIInputs{Args: []string{"install", strings.Repeat("a", maxArgLen)}},
			wantPass: true,
		},
		{
			name:       "exceeded arg length rejected (4097 chars)",
			inputs:     CLIInputs{Args: []string{"install", strings.Repeat("b", maxArgLen+1)}},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: fmt.Sprintf("argument 1 length %d exceeds maximum limit (%d)", maxArgLen+1, maxArgLen),
		},
		{
			name:       "null byte injection in command rejected",
			inputs:     CLIInputs{Args: []string{"pip\x00evil"}},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: "contains forbidden null byte",
		},
		{
			name:       "null byte injection in argument rejected",
			inputs:     CLIInputs{Args: []string{"install", "pkg\x00injection"}},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: "contains forbidden null byte",
		},

		// UVOverride schema boundaries
		{
			name:     "valid absolute path override accepted",
			inputs:   CLIInputs{Args: []string{"list"}, UVOverride: validAbsPath},
			wantPass: true,
		},
		{
			name:       "blank override rejected",
			inputs:     CLIInputs{Args: []string{"list"}, UVOverride: "   "},
			wantPass:   false,
			wantTyped:  ErrInvalidOverride,
			wantErrSub: "UVPIP_UV is set but blank",
		},
		{
			name:       "whitespace padded override rejected",
			inputs:     CLIInputs{Args: []string{"list"}, UVOverride: " " + validAbsPath + " "},
			wantPass:   false,
			wantTyped:  ErrInvalidOverride,
			wantErrSub: "leading or trailing whitespace",
		},
		{
			name:       "relative path override rejected",
			inputs:     CLIInputs{Args: []string{"list"}, UVOverride: "custom/bin/uv"},
			wantPass:   false,
			wantTyped:  ErrInvalidOverride,
			wantErrSub: "must be an absolute executable path",
		},

		// Combined boundary assertions
		{
			name:       "invalid args take precedence when override also evaluated",
			inputs:     CLIInputs{Args: []string{""}, UVOverride: validAbsPath},
			wantPass:   false,
			wantTyped:  ErrInvalidArgs,
			wantErrSub: "pip command is empty",
		},
		{
			name:       "invalid override fails when args valid",
			inputs:     CLIInputs{Args: []string{"freeze"}, UVOverride: "relative/uv"},
			wantPass:   false,
			wantTyped:  ErrInvalidOverride,
			wantErrSub: "must be an absolute executable path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputs(tt.inputs)
			if tt.wantPass {
				if err != nil {
					t.Fatalf("expected valid inputs to pass, got error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected failure for %s, but got nil", tt.name)
			}
			if tt.wantTyped != nil && !errors.Is(err, tt.wantTyped) {
				t.Fatalf("expected error wrapping %v, got: %v", tt.wantTyped, err)
			}
			if tt.wantErrSub != "" && !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Fatalf("expected error containing %q, got: %v", tt.wantErrSub, err)
			}
		})
	}
}

func makeArgList(prefix string, count int) []string {
	args := make([]string, count)
	args[0] = "install"
	for i := 1; i < count; i++ {
		args[i] = fmt.Sprintf("%s-%d", prefix, i)
	}
	return args
}

// Python selectors are forwarded to uv after the normal argument boundary checks.
func TestValidatePythonArgumentBoundary(t *testing.T) {
	for _, tt := range []struct {
		name, selector string
		wantErr        bool
	}{
		{"empty selector forwarded", "", false},
		{"interpreter name", "python3", false},
		{"absolute interpreter path", "/usr/bin/python3", false},
		{"path with spaces", "/tools/my python", false},
		{"null byte rejected", "/usr/bin/python3\x00malicious", true},
		{"byte limit accepted", strings.Repeat("a", maxArgLen), false},
		{"byte limit exceeded", strings.Repeat("a", maxArgLen+1), true},
		{"multibyte limit exceeded", strings.Repeat("é", maxArgLen/2+1), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputs(CLIInputs{Args: []string{"install", "--python", tt.selector, "pkg"}})
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, ErrInvalidArgs)) {
				t.Fatalf("validation error = %v, want rejection %v", err, tt.wantErr)
			}
		})
	}
}
