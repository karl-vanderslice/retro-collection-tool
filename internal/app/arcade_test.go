package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
)

func TestSelectArcadeEntriesFiltersAndKeepsBios(t *testing.T) {
	t.Parallel()

	entries := []arcadeDATEntry{
		{Name: "sf2", Description: "Street Fighter II"},
		{Name: "sf2j", Description: "Street Fighter II (Japan)", CloneOf: "sf2"},
		{Name: "mjgame", Description: "Super Mahjong Deluxe"},
		{Name: "neogeo", Description: "Neo Geo Bios", IsBios: true},
	}

	sel := selectArcadeEntries(entries, []string{"mahjong"})
	if len(sel.Games) != 1 || sel.Games[0] != "sf2" {
		t.Fatalf("unexpected game selection: %#v", sel.Games)
	}
	if len(sel.Bios) != 1 || sel.Bios[0] != "neogeo" {
		t.Fatalf("unexpected bios selection: %#v", sel.Bios)
	}
}

func TestVerifyArcadeVaultSetReportsGamesAndBios(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfg := testArcadeConfig(tmp)
	specs := arcadeSpecsFromConfig(cfg)

	if err := writeTestDAT(specs[0].DatPath, []string{
		`<machine name="sf2"><description>Street Fighter II</description></machine>`,
		`<machine name="sf2j" cloneof="sf2"><description>Street Fighter II (Japan)</description></machine>`,
		`<machine name="mjgame"><description>Mahjong Carnival</description></machine>`,
		`<machine name="neogeo" isbios="yes"><description>Neo Geo Bios</description></machine>`,
	}); err != nil {
		t.Fatalf("write mame dat: %v", err)
	}
	if err := os.MkdirAll(specs[0].VaultDir, 0o755); err != nil {
		t.Fatalf("mkdir vault: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specs[0].VaultDir, "sf2.zip"), []byte("rom"), 0o644); err != nil {
		t.Fatalf("write game: %v", err)
	}

	report, err := verifyArcadeVaultSet(specs[0], cfg)
	if err != nil {
		t.Fatalf("verifyArcadeVaultSet: %v", err)
	}
	if report.TotalGames != 1 || report.PresentGames != 1 || report.MissingGames != 0 {
		t.Fatalf("unexpected game report: %#v", report)
	}
	if report.TotalBios != 1 || report.PresentBios != 0 || report.MissingBios != 1 {
		t.Fatalf("unexpected bios report: %#v", report)
	}
}

func TestRunArcadeSyncLinksGamesAndBios(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfg := testArcadeConfig(tmp)
	specs := arcadeSpecsFromConfig(cfg)

	if err := writeTestDAT(specs[0].DatPath, []string{
		`<machine name="sf2"><description>Street Fighter II</description></machine>`,
		`<machine name="sf2j" cloneof="sf2"><description>Street Fighter II (Japan)</description></machine>`,
		`<machine name="neogeo" isbios="yes"><description>Neo Geo Bios</description></machine>`,
	}); err != nil {
		t.Fatalf("write mame dat: %v", err)
	}
	if err := writeTestDAT(specs[1].DatPath, []string{
		`<game name="kof98"><description>King of Fighters '98</description></game>`,
		`<game name="neogeo" isbios="yes"><description>Neo Geo Bios</description></game>`,
	}); err != nil {
		t.Fatalf("write fbneo dat: %v", err)
	}

	if err := os.MkdirAll(specs[0].VaultDir, 0o755); err != nil {
		t.Fatalf("mkdir mame vault: %v", err)
	}
	if err := os.MkdirAll(specs[1].VaultDir, 0o755); err != nil {
		t.Fatalf("mkdir fbneo vault: %v", err)
	}

	mameGame := filepath.Join(specs[0].VaultDir, "sf2.zip")
	mameBios := filepath.Join(specs[0].VaultDir, "neogeo.zip")
	fbGame := filepath.Join(specs[1].VaultDir, "kof98.zip")
	fbBios := filepath.Join(specs[1].VaultDir, "neogeo.zip")

	for _, p := range []string{mameGame, mameBios, fbGame, fbBios} {
		if err := os.WriteFile(p, []byte("rom"), 0o644); err != nil {
			t.Fatalf("write vault file %s: %v", p, err)
		}
	}

	if err := runArcade(cfg, globalFlags{}, []string{"sync"}); err != nil {
		t.Fatalf("runArcade sync: %v", err)
	}

	assertHardLinked(t, mameGame, filepath.Join(specs[0].LibraryDir, "sf2.zip"))
	assertHardLinked(t, mameBios, filepath.Join(specs[0].LibraryDir, "neogeo.zip"))
	assertHardLinked(t, fbGame, filepath.Join(specs[1].LibraryDir, "kof98.zip"))
	assertHardLinked(t, fbBios, filepath.Join(specs[1].LibraryDir, "neogeo.zip"))

	if _, err := os.Stat(filepath.Join(specs[0].LibraryDir, "sf2j.zip")); !os.IsNotExist(err) {
		t.Fatalf("expected clone not to be linked")
	}
}

func TestRunArcadeDatsUpdateDownloads(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mame.dat":
			_, _ = w.Write([]byte("<datafile><machine name=\"sf2\"><description>Street Fighter II</description></machine></datafile>"))
		case "/fbneo.dat":
			_, _ = w.Write([]byte("<datafile><game name=\"kof98\"><description>KOF</description></game></datafile>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := testArcadeConfig(tmp)
	cfg.Arcade.DatMAME2003URL = server.URL + "/mame.dat"
	cfg.Arcade.DatFBNeoURL = server.URL + "/fbneo.dat"

	if err := runArcade(cfg, globalFlags{}, []string{"dats", "update"}); err != nil {
		t.Fatalf("runArcade dats update: %v", err)
	}

	specs := arcadeSpecsFromConfig(cfg)
	for _, spec := range specs {
		if _, err := os.Stat(spec.DatPath); err != nil {
			t.Fatalf("expected dat at %s: %v", spec.DatPath, err)
		}
	}
}

func testArcadeConfig(root string) *config.Config {
	return &config.Config{
		Root:     root,
		CacheDir: "cache",
		Paths: config.PathsConfig{
			RommLibraryRoms: "roms/Library/roms",
		},
		Features: config.FeatureToggles{
			EnableArcade: true,
		},
		Arcade: config.ArcadeConfig{
			VaultMAME2003Plus: "roms/Vault/Arcade/mame-2003-plus-reference-set/roms",
			VaultFBNeo:        "roms/Vault/Arcade/fbneo_1003_bestset/fbneo_1_0_0_3_best/games",
			LibraryMAME2003:   "roms/Library/roms/arcade/mame-2003-plus",
			LibraryFBNeo:      "roms/Library/roms/arcade/fbneo",
			DatMAME2003File:   "arcade-mame-2003-plus.dat",
			DatFBNeoFile:      "arcade-fbneo.dat",
			ExcludeKeywords:   []string{"mahjong"},
		},
	}
}

func writeTestDAT(path string, nodes []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	xml := "<datafile>" + strings.Join(nodes, "") + "</datafile>"
	return os.WriteFile(path, []byte(xml), 0o644)
}

func assertHardLinked(t *testing.T, src, dst string) {
	t.Helper()
	srcInfo, err := os.Stat(src)
	if err != nil {
		t.Fatalf("stat src %s: %v", src, err)
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst %s: %v", dst, err)
	}
	if !os.SameFile(srcInfo, dstInfo) {
		t.Fatalf("expected hard link between %s and %s", src, dst)
	}
}

func TestRunArcadeDatsVerifyRequiresExistingDats(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfg := testArcadeConfig(tmp)
	err := runArcade(cfg, globalFlags{}, []string{"dats", "verify"})
	if err == nil {
		t.Fatalf("expected missing dat error")
	}
	if got := err.Error(); got == "" {
		t.Fatalf("expected non-empty error")
	}
	if !strings.Contains(err.Error(), "run: retro-collection-tool arcade dats update") {
		t.Fatalf("expected update guidance in error: %v", err)
	}
}
