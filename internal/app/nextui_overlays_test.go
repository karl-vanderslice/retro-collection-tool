package app

import "testing"

func TestParseNextUIOverlayProviders(t *testing.T) {
	t.Parallel()

	got, err := parseNextUIOverlayProviders(" krutzotrem , skywalker541, skywalker ")
	if err != nil {
		t.Fatalf("parseNextUIOverlayProviders: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 unique providers, got %d", len(got))
	}
	if got[0] != overlayProviderKrutzotrem || got[1] != overlayProviderSkywalker {
		t.Fatalf("unexpected providers: %#v", got)
	}
}

func TestParseNextUIOverlayProvidersRejectsUnknown(t *testing.T) {
	t.Parallel()

	if _, err := parseNextUIOverlayProviders("unknown-provider"); err == nil {
		t.Fatal("expected parseNextUIOverlayProviders to reject unknown provider")
	}
}

func TestMapOverlayEntryPathFromOverlaysRoot(t *testing.T) {
	t.Parallel()

	index := map[string]string{"GB": "GB", "GBC": "GBC"}
	got, ok := mapOverlayEntryPath("Overlays/GB/BRICK_GB_NATIVE.png", index)
	if !ok {
		t.Fatal("expected overlay path to map")
	}
	if want := "Overlays/GB/BRICK_GB_NATIVE.png"; got != want {
		t.Fatalf("mapOverlayEntryPath mismatch: got %q want %q", got, want)
	}
}

func TestMapOverlayEntryPathFromRepoRoot(t *testing.T) {
	t.Parallel()

	index := map[string]string{"GB": "GB", "GBC": "GBC"}
	got, ok := mapOverlayEntryPath("GBC/STN (Authentic Look)/BRICK_GBC_NATIVE_COOL_STN.png", index)
	if !ok {
		t.Fatal("expected repo-root path to map")
	}
	if want := "Overlays/GBC/STN (Authentic Look)/BRICK_GBC_NATIVE_COOL_STN.png"; got != want {
		t.Fatalf("mapOverlayEntryPath mismatch: got %q want %q", got, want)
	}
}

func TestMapOverlayEntryPathSkipsUnknownFolder(t *testing.T) {
	t.Parallel()

	index := map[string]string{"GB": "GB"}
	if _, ok := mapOverlayEntryPath("README.md", index); ok {
		t.Fatal("expected file without system folder to be skipped")
	}
	if _, ok := mapOverlayEntryPath("GBA/overlay.png", index); ok {
		t.Fatal("expected unknown overlay folder to be skipped")
	}
}
