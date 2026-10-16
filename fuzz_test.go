package main

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func FuzzTranslateToUV(f *testing.F) {
	f.Add("install", "requests", "-y")
	f.Add("uninstall", "pkg", "--yes")
	f.Add("upgrade", "mypkg", "")
	f.Add("cache", "dir", "")
	f.Add("", "", "")
	f.Add("uninstall", "--", "--yes")
	f.Fuzz(func(t *testing.T, cmd, arg1, arg2 string) {
		args := []string{cmd, arg1, arg2}
		out := translateToUV(args)
		if len(out) == 0 {
			t.Fatal("expected non-empty translated arguments")
		}
		if out[0] != "pip" && out[0] != "cache" {
			t.Fatalf("unexpected routing prefix: %q", out[0])
		}
		// The option terminator makes the final argument a literal operand.
		if arg1 == "--" && out[len(out)-1] != arg2 {
			t.Fatal("literal operand after -- was changed")
		}
		out[0] = "mutated"
		if !reflect.DeepEqual(args, []string{cmd, arg1, arg2}) {
			t.Fatal("translation mutated or aliased the input")
		}
	})
}

func FuzzValidateUVOverride(f *testing.F) {
	f.Add("/usr/bin/uv")
	f.Add("C:\\uv\\uv.exe")
	f.Add("  /usr/bin/uv  ")
	f.Add("relative/path")
	f.Add("")
	f.Fuzz(func(t *testing.T, path string) {
		err := validateUVOverride(path)
		if err != nil && !errors.Is(err, ErrInvalidOverride) {
			t.Fatalf("override rejection lost its type: %v", err)
		}
		if err == nil && (!filepath.IsAbs(path) || strings.TrimSpace(path) != path) {
			t.Fatal("accepted a relative or whitespace-padded override")
		}
	})
}

func FuzzBuildEnv(f *testing.F) {
	f.Add("PATH=/bin", "VIRTUAL_ENV=/venv", "UV_SYSTEM_PYTHON=1")
	f.Add("CONDA_PREFIX=/conda", "FOO=BAR", "")
	f.Add("", "", "")
	f.Fuzz(func(t *testing.T, e1, e2, e3 string) {
		env := []string{e1, e2, e3}
		out := buildEnv(env)
		if len(out) < len(env) || len(out) > len(env)+1 || !reflect.DeepEqual(out[:len(env)], env) {
			t.Fatal("buildEnv changed existing environment entries")
		}
		if len(out) > len(env) && out[len(env)] != "UV_SYSTEM_PYTHON=1" {
			t.Fatal("buildEnv appended an unexpected setting")
		}
		if !reflect.DeepEqual(buildEnv(out), out) {
			t.Fatal("environment defaulting is not idempotent")
		}
		out[0] = "mutated"
		if !reflect.DeepEqual(env, []string{e1, e2, e3}) {
			t.Fatal("buildEnv mutated or aliased the input")
		}
	})
}

func FuzzValidateArgBounds(f *testing.F) {
	for _, seed := range []struct {
		value string
		count uint16
	}{
		{"pkg", 0}, {"pkg", 1}, {"pkg", 512}, {"pkg", 513},
		{strings.Repeat("a", 4096), 1}, {strings.Repeat("a", 4097), 1},
		{"pkg\x00name", 1}, {"é", 512},
	} {
		f.Add(seed.value, seed.count)
	}
	f.Fuzz(func(t *testing.T, value string, count uint16) {
		// Bound allocation while exercising both sides of the 512-argument limit.
		args := make([]string, int(count)%514)
		for i := range args {
			args[i] = value
		}
		err := validateArgBounds(args)
		invalid := len(args) > 512 || (len(args) > 0 && (len(value) > 4096 || strings.ContainsRune(value, '\x00')))
		if (err != nil) != invalid {
			t.Fatalf("boundary acceptance mismatch: count=%d bytes=%d error=%v", len(args), len(value), err)
		}
		if err != nil && (!errors.Is(err, ErrInvalidArgs) || !errors.Is(err, errUsage)) {
			t.Fatalf("boundary rejection lost its error type: %v", err)
		}
		for _, arg := range args {
			if arg != value {
				t.Fatal("validation mutated input")
			}
		}
	})
}
