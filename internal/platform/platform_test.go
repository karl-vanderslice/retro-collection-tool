package platform

import (
	"reflect"
	"testing"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
)

func TestExpandSystemsResolvesNeoGeoPocketAliases(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Systems: map[string]config.SystemConfig{
			"neo-geo-pocket": {
				Enabled: true,
			},
			"neo-geo-pocket-color": {
				Enabled: true,
			},
		},
	}

	got, err := ExpandSystems([]string{"ngpc,ngp"}, false, cfg)
	if err != nil {
		t.Fatalf("ExpandSystems: %v", err)
	}

	want := []string{"neo-geo-pocket", "neo-geo-pocket-color"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected systems: got %#v want %#v", got, want)
	}
}

func TestExpandSystemsResolvesNaturalNameVariant(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Systems: map[string]config.SystemConfig{
			"neo-geo-pocket-color": {
				Enabled: true,
			},
		},
	}

	got, err := ExpandSystems([]string{"neo geo pocket color"}, false, cfg)
	if err != nil {
		t.Fatalf("ExpandSystems: %v", err)
	}

	want := []string{"neo-geo-pocket-color"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected systems: got %#v want %#v", got, want)
	}
}

func TestExpandSystemsDedupesAliasAndCanonical(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Systems: map[string]config.SystemConfig{
			"neo-geo-pocket-color": {
				Enabled: true,
			},
		},
	}

	got, err := ExpandSystems([]string{"neo-geo-pocket-color,ngpc"}, false, cfg)
	if err != nil {
		t.Fatalf("ExpandSystems: %v", err)
	}

	want := []string{"neo-geo-pocket-color"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected systems: got %#v want %#v", got, want)
	}
}
