package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// errUsage marks input rejected before uv starts; runCLI maps it to exit status 2.
// It wraps ErrInvalidArgs so callers can inspect failure types programmatically.
var errUsage = fmt.Errorf("%w: usage error", ErrInvalidArgs)

// CLIInputs specifies all user-supplied inputs evaluated at CLI boundaries:
// command arguments and environment configuration overrides.
type CLIInputs struct {
	Args       []string
	UVOverride string
}

// ValidateInputs runs boundary validation on all CLI entry point inputs.
// It returns a typed error wrapping ErrInvalidArgs if arguments are invalid,
// or ErrInvalidOverride if environment configuration overrides are malformed.
func ValidateInputs(inputs CLIInputs) error {
	if len(inputs.Args) > 0 || inputs.UVOverride == "" {
		if err := validateArgs(inputs.Args); err != nil {
			return err
		}
	}
	if inputs.UVOverride != "" {
		if err := validateUVOverride(inputs.UVOverride); err != nil {
			return err
		}
	}
	return nil
}

const (
	maxArgCount = 512
	maxArgLen   = 4096
)

// validateArgBounds verifies schema boundary limits on argument count, argument length, and null-byte injection.
func validateArgBounds(args []string) error {
	if len(args) > maxArgCount {
		return fmt.Errorf("%w: argument count %d exceeds maximum limit (%d)", errUsage, len(args), maxArgCount)
	}
	for i, a := range args {
		if len(a) > maxArgLen {
			return fmt.Errorf("%w: argument %d length %d exceeds maximum limit (%d)", errUsage, i, len(a), maxArgLen)
		}
		if strings.ContainsRune(a, '\x00') {
			return fmt.Errorf("%w: argument %d contains forbidden null byte", errUsage, i)
		}
	}
	return nil
}

// validateArgs rejects a blank pip command, which uv would only report as an
// unrecognized empty subcommand, and enforces argument boundary constraints.
// Values are never echoed: arguments can embed index credentials.
func validateArgs(args []string) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("%w: pip command is empty", errUsage)
	}
	if err := validateArgBounds(args); err != nil {
		return err
	}
	return validateTranslationFlags(args)
}

// validateUVOverride checks a non-empty UVPIP_UV before path lookup, so common
// quoting mistakes get a precise message instead of "executable not found".
func validateUVOverride(path string) error {
	switch {
	case strings.TrimSpace(path) == "":
		return fmt.Errorf("%w: UVPIP_UV is set but blank; unset it or give an absolute path", ErrInvalidOverride)
	case strings.TrimSpace(path) != path:
		return fmt.Errorf("%w: UVPIP_UV has leading or trailing whitespace; check its quoting", ErrInvalidOverride)
	case !filepath.IsAbs(path):
		return fmt.Errorf("%w: UVPIP_UV must be an absolute executable path", ErrInvalidOverride)
	}
	return nil
}
