package app

import (
	"errors"
	"strings"
	"testing"
)

func TestFormatCLIErrorNoCommandIncludesHelpHint(t *testing.T) {
	t.Parallel()

	got := FormatCLIError(errors.New("no command provided"))
	if !strings.Contains(got, "retro-collection-tool help") {
		t.Fatalf("expected help hint, got %q", got)
	}
}

func TestFormatCLIErrorUnknownCommandIncludesDiscoveryHints(t *testing.T) {
	t.Parallel()

	got := FormatCLIError(errors.New("unknown command: sycn (did you mean \"sync\"?)"))
	if !strings.Contains(got, "Check available commands") {
		t.Fatalf("expected discovery hint, got %q", got)
	}
	if !strings.Contains(got, "Command names are lowercase") {
		t.Fatalf("expected casing hint, got %q", got)
	}
}

func TestFormatCLIErrorExportDestinationHint(t *testing.T) {
	t.Parallel()

	got := FormatCLIError(errors.New("export requires --destination"))
	if !strings.Contains(got, "--destination /media/SDCARD") {
		t.Fatalf("expected destination example, got %q", got)
	}
}

func TestFormatCLIErrorReturnsOriginalWhenNoHints(t *testing.T) {
	t.Parallel()

	msg := "some low-level failure"
	got := FormatCLIError(errors.New(msg))
	if got != msg {
		t.Fatalf("expected original message, got %q", got)
	}
}
