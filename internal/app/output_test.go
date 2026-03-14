package app

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseGlobalFlagsNoColorShorthand(t *testing.T) {
	t.Parallel()

	g, rest, err := parseGlobalFlags([]string{"--no-color", "sync"})
	if err != nil {
		t.Fatalf("parseGlobalFlags: %v", err)
	}
	if g.colorMode != colorModeNever {
		t.Fatalf("expected color mode %q, got %q", colorModeNever, g.colorMode)
	}
	if len(rest) != 1 || rest[0] != "sync" {
		t.Fatalf("unexpected args: %#v", rest)
	}
}

func TestParseGlobalFlagsRejectsInvalidColorMode(t *testing.T) {
	t.Parallel()

	_, _, err := parseGlobalFlags([]string{"--color", "rainbow", "sync"})
	if err == nil {
		t.Fatal("expected error for invalid color mode")
	}
}

func TestUsesColorAlwaysModeRespectsNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	g := globalFlags{outputMode: outputModeHuman, colorMode: colorModeAlways}
	if g.usesColor() {
		t.Fatal("expected NO_COLOR to disable color output")
	}
}

func TestEmitLineHumanNoColorFormatting(t *testing.T) {
	g := globalFlags{outputMode: outputModeHuman, colorMode: colorModeNever}
	out := captureStdout(t, func() {
		emitInfo(g, "sync", "retail", "accepted", outputFields{"systems": "snes", "dry_run": true})
	})
	out = strings.TrimSpace(out)
	if !strings.Contains(out, "[SYNC] [RETAIL] [INFO] accepted") {
		t.Fatalf("unexpected prefix: %q", out)
	}
	if !strings.Contains(out, "dry_run=true") || !strings.Contains(out, "systems=snes") {
		t.Fatalf("expected formatted fields, got %q", out)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return buf.String()
}
