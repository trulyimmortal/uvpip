package main

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidArgs indicates invalid or empty CLI arguments.
	ErrInvalidArgs = errors.New("invalid arguments")

	// ErrUsage indicates CLI argument or flag usage violations (exit status 2).
	ErrUsage = fmt.Errorf("%w: usage error", ErrInvalidArgs)

	// ErrInvalidOverride indicates malformed UVPIP_UV environment variable configuration.
	ErrInvalidOverride = errors.New("invalid uv override")

	// ErrTranslation indicates command translation failure.
	ErrTranslation = errors.New("command translation failed")

	// ErrUVNotFound indicates uv could not be located in PATH or candidate locations.
	ErrUVNotFound = errors.New("uv executable not found")

	// ErrCannotExecute indicates uv binary was found but cannot be started.
	ErrCannotExecute = errors.New("cannot execute uv binary")

	// ErrRecursiveExecution prevents uvpip from calling itself in an infinite loop.
	ErrRecursiveExecution = errors.New("uv resolves to uvpip itself; refusing recursive execution")
)
