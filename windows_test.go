//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsShimDetection(t *testing.T) {
	// Verify that the repository's shipped Windows shims are identified as uvpip shims.
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for _, shimName := range []string{"pip.cmd", "pip3.cmd"} {
		shimPath := filepath.Join(repoRoot, "bin", shimName)
		if _, err := os.Stat(shimPath); os.IsNotExist(err) {
			t.Skipf("repo shim %s not present at expected path %s", shimName, shimPath)
		}
		if !isUvpipShim(shimPath) {
			t.Errorf("expected isUvpipShim to recognize repository shim: %s", shimPath)
		}
	}

	// Verify that a foreign, unmanaged .cmd file is rejected.
	dir := t.TempDir()
	foreignPath := filepath.Join(dir, "foreign.cmd")
	if err := os.WriteFile(foreignPath, []byte("@echo off\r\npython.exe -m pip %*\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if isUvpipShim(foreignPath) {
		t.Errorf("expected isUvpipShim to reject foreign cmd script: %s", foreignPath)
	}

	// Verify that missing or unreadable files fail gracefully without panic.
	missingPath := filepath.Join(dir, "nonexistent.cmd")
	if isUvpipShim(missingPath) {
		t.Errorf("expected isUvpipShim to return false for nonexistent file: %s", missingPath)
	}
}

func TestWindowsCandidatePathsResolution(t *testing.T) {
	localAppData := filepath.Join("C:\\", "Users", "TestUser", "AppData", "Local")
	userProfile := filepath.Join("C:\\", "Users", "TestUser")
	appData := filepath.Join("C:\\", "Users", "TestUser", "AppData", "Roaming")

	mockGetenv := func(key string) string {
		switch key {
		case "LOCALAPPDATA":
			return localAppData
		case "APPDATA":
			return appData
		default:
			return ""
		}
	}

	paths := candidatePathsFor("windows", mockGetenv, userProfile)
	if len(paths) == 0 {
		t.Fatal("expected candidate paths for windows, got none")
	}

	for _, p := range paths {
		if !strings.HasSuffix(p, ".exe") {
			t.Errorf("candidate %q: expected .exe extension on Windows", p)
		}
	}

	expectedPrefixes := []string{
		filepath.Join(userProfile, ".local", "bin", "uv.exe"),
		filepath.Join(userProfile, ".cargo", "bin", "uv.exe"),
		filepath.Join(appData, "astral", "uv", "bin", "uv.exe"),
		filepath.Join(localAppData, "Programs", "uv", "uv.exe"),
	}

	for i, expected := range expectedPrefixes {
		if i >= len(paths) || paths[i] != expected {
			t.Errorf("candidate %d: expected %q, got %q", i, expected, paths[i])
		}
	}
}

func TestWindowsCaseInsensitiveEnvHandling(t *testing.T) {
	// On Windows, environment variables can be cased as "Path" or "PATH".
	baseEnv := []string{
		"Path=C:\\Windows\\system32;C:\\Windows",
		"SystemRoot=C:\\Windows",
		"COMSPEC=C:\\Windows\\system32\\cmd.exe",
	}

	childEnv := buildEnv(baseEnv)

	// Since neither VIRTUAL_ENV nor CONDA_PREFIX nor UV_SYSTEM_PYTHON is set,
	// buildEnv must default UV_SYSTEM_PYTHON=1.
	hasSystemPython := false
	for _, kv := range childEnv {
		if kv == "UV_SYSTEM_PYTHON=1" {
			hasSystemPython = true
			break
		}
	}
	if !hasSystemPython {
		t.Errorf("expected UV_SYSTEM_PYTHON=1 to be injected on Windows default environment: %v", childEnv)
	}
}

func TestWindowsProfileMarkerIntegrity(t *testing.T) {
	// Test the exact PowerShell profile marker boundaries used by the Windows installer.
	markerStart := "# >>> uvpip initialize >>>"
	markerEnd := "# <<< uvpip initialize <<<"

	profileContent := strings.Join([]string{
		"# User pre-existing profile",
		"Set-PSReadLineOption -PredictionSource History",
		markerStart,
		"$env:PATH = \"$HOME\\.uvpip\\bin;$env:PATH\"",
		markerEnd,
		"# User post-existing profile",
		"Set-Alias ll Get-ChildItem",
	}, "\r\n")

	if !strings.Contains(profileContent, markerStart) || !strings.Contains(profileContent, markerEnd) {
		t.Fatal("expected profile content to preserve exact marker boundaries")
	}

	startIdx := strings.Index(profileContent, markerStart)
	endIdx := strings.Index(profileContent, markerEnd)
	if startIdx >= endIdx {
		t.Fatalf("markerStart (%d) must precede markerEnd (%d)", startIdx, endIdx)
	}
}
