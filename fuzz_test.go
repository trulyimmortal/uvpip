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
