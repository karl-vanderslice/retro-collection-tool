package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextUISystemFolderName(t *testing.T) {
	t.Parallel()

	if got := nextUISystemFolderName("GBA"); got != "16) Game Boy Advance (GBA)" {
		t.Fatalf("unexpected nextui folder: %q", got)
	}
	if got := nextUISystemFolderName("Nintendo Entertainment System (FC)"); got != "Nintendo Entertainment System (FC)" {
		t.Fatalf("expected folder with tag to pass through, got %q", got)
	}
}

func TestCopySystemROMsFlattenSkipsManualsAndImgs(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, "root-game.nes"), "rom")
	mustWriteFile(t, filepath.Join(src, "All but the Best (NES)", "sub-game.nes"), "rom")
	mustWriteFile(t, filepath.Join(src, "Manuals", "manual.pdf"), "doc")
	mustWriteFile(t, filepath.Join(src, "Imgs", "root-game.png"), "art")

	copied, duplicates, converted, _, err := copySystemROMs("FC", src, dst, systemModeFlatten, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs: %v", err)
	}
	if copied != 2 {
		t.Fatalf("expected 2 copied ROMs, got %d", copied)
	}
	if duplicates != 0 {
		t.Fatalf("expected 0 duplicates, got %d", duplicates)
	}
	if converted != 0 {
		t.Fatalf("expected 0 7z conversions, got %d", converted)
	}

	assertFileExists(t, filepath.Join(dst, "root-game.nes"))
	assertFileExists(t, filepath.Join(dst, "sub-game.nes"))
	assertFileMissing(t, filepath.Join(dst, "manual.pdf"))
	assertFileMissing(t, filepath.Join(dst, "root-game.png"))
}

func TestCopySystemROMsFlattenSkipsPuristCategories(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, "All but the Best (MD)", "Core Game.zip"), "rom")
	mustWriteFile(t, filepath.Join(src, "Hacks (MD)", "Hack One.zip"), "hack")
	mustWriteFile(t, filepath.Join(src, "Translations (MD)", "Translation One.zip"), "translation")
	mustWriteFile(t, filepath.Join(src, "Unlicensed Homebrew (MD)", "Homebrew One.zip"), "homebrew")

	copied, duplicates, converted, _, err := copySystemROMs("MD", src, dst, systemModeFlatten, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs purist categories: %v", err)
	}
	if copied != 1 || duplicates != 0 || converted != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}

	assertFileExists(t, filepath.Join(dst, "Core Game.zip"))
	assertFileMissing(t, filepath.Join(dst, "Hack One.zip"))
	assertFileMissing(t, filepath.Join(dst, "Translation One.zip"))
	assertFileMissing(t, filepath.Join(dst, "Homebrew One.zip"))
}

func TestCopySystemROMsMDRoutes32XToDedicatedFolder(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	romsRoot := t.TempDir()
	mdDst := filepath.Join(romsRoot, nextUISystemFolderName("MD"))

	mustWriteFile(t, filepath.Join(src, "32X Games (Genesis)", "Doom.zip"), "rom")
	mustWriteFile(t, filepath.Join(src, "All but the Best (Genesis)", "Sonic.zip"), "rom")

	copied, duplicates, converted, _, err := copySystemROMs("MD", src, mdDst, systemModeFlatten, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs md 32x route: %v", err)
	}
	if copied != 2 || duplicates != 0 || converted != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}

	assertFileExists(t, filepath.Join(mdDst, "Sonic.zip"))
	assertFileExists(t, filepath.Join(romsRoot, md32XSystemFolderName, "Doom.zip"))
	assertFileMissing(t, filepath.Join(mdDst, "Doom.zip"))
}

func TestCopySystemROMsTreePreservesSubdirsForScummVM(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, "Game One", "data.001"), "bin")
	mustWriteFile(t, filepath.Join(src, "Imgs", "cover.png"), "art")

	copied, duplicates, converted, _, err := copySystemROMs("SCUMMVM", src, dst, systemModeTree, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs tree: %v", err)
	}
	if copied != 1 || duplicates != 0 || converted != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}

	assertFileExists(t, filepath.Join(dst, "Game One", "data.001"))
	assertFileMissing(t, filepath.Join(dst, "cover.png"))
}

func TestCopySystemROMsFlattenPreservesHiddenDiscFiles(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, ".hidden", "Disc 1.chd"), "chd")
	mustWriteFile(t, filepath.Join(src, "Game.m3u"), ".hidden/Disc 1.chd\n")

	copied, duplicates, converted, _, err := copySystemROMs("PS", src, dst, systemModeFlatten, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs flatten hidden: %v", err)
	}
	if copied != 2 || duplicates != 0 || converted != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}

	assertFileExists(t, filepath.Join(dst, ".hidden", "Disc 1.chd"))
	assertFileExists(t, filepath.Join(dst, "Game.m3u"))
}

func TestCopySystemROMsArcadePreservesCHDRelativePath(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, "kinst", "kinst.chd"), "chd")
	mustWriteFile(t, filepath.Join(src, "kinst.zip"), "zip")

	copied, duplicates, converted, _, err := copySystemROMs("ARCADE", src, dst, systemModeFlatten, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs arcade chd: %v", err)
	}
	if copied != 2 || duplicates != 0 || converted != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}

	assertFileExists(t, filepath.Join(dst, "kinst", "kinst.chd"))
	assertFileExists(t, filepath.Join(dst, "kinst.zip"))
}

func TestCopySystemROMsDryRunCounts7zConversion(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, "pack.7z"), "placeholder")

	copied, duplicates, converted, _, err := copySystemROMs("FC", src, dst, systemModeFlatten, false, globalFlags{dryRun: true})
	if err != nil {
		t.Fatalf("copySystemROMs dry-run 7z: %v", err)
	}
	if copied != 1 || duplicates != 0 || converted != 1 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}
}

func TestCopySystemROMsFlattenSkipsCFGSidecar(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, "Game.tap"), "rom")
	mustWriteFile(t, filepath.Join(src, "Game.tap.p2k.cfg"), "cfg")

	copied, duplicates, converted, _, err := copySystemROMs("ZXS", src, dst, systemModeFlatten, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs flatten cfg: %v", err)
	}
	if copied != 1 || duplicates != 0 || converted != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}

	assertFileExists(t, filepath.Join(dst, "Game.tap"))
	assertFileMissing(t, filepath.Join(dst, "Game.tap.p2k.cfg"))
}

func TestCopySystemROMsSkipsCommonCruftFiles(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	mustWriteFile(t, filepath.Join(src, "Game.nes"), "rom")
	mustWriteFile(t, filepath.Join(src, ".DS_Store"), "junk")
	mustWriteFile(t, filepath.Join(src, "Thumbs.db"), "junk")
	mustWriteFile(t, filepath.Join(src, "desktop.ini"), "junk")
	mustWriteFile(t, filepath.Join(src, "__MACOSX", "artifact.bin"), "junk")

	copied, duplicates, converted, _, err := copySystemROMs("FC", src, dst, systemModeFlatten, false, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemROMs cruft: %v", err)
	}
	if copied < 1 || duplicates != 0 || converted != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d converted=%d", copied, duplicates, converted)
	}

	assertFileExists(t, filepath.Join(dst, "Game.nes"))
	assertFileMissing(t, filepath.Join(dst, ".DS_Store"))
	assertFileMissing(t, filepath.Join(dst, "Thumbs.db"))
	assertFileMissing(t, filepath.Join(dst, "desktop.ini"))
	assertFileMissing(t, filepath.Join(dst, "artifact.bin"))
}

func TestNormalizeSystemZipUniformityDryRun(t *testing.T) {
	t.Parallel()

	systemDir := t.TempDir()
	mustWriteFile(t, filepath.Join(systemDir, "A.zip"), "zip")
	mustWriteFile(t, filepath.Join(systemDir, "B.zip"), "zip")
	mustWriteFile(t, filepath.Join(systemDir, "C.gba"), "raw")

	converted, err := normalizeSystemZipUniformity(systemDir, systemModeFlatten, globalFlags{dryRun: true})
	if err != nil {
		t.Fatalf("normalizeSystemZipUniformity dry-run: %v", err)
	}
	if converted != 1 {
		t.Fatalf("expected 1 dry-run conversion, got %d", converted)
	}
	assertFileExists(t, filepath.Join(systemDir, "C.gba"))
}

func TestCopySystemArtworkCopiesPNGsOnly(t *testing.T) {
	t.Parallel()

	srcSystem := t.TempDir()
	dstMedia := filepath.Join(t.TempDir(), ".media")
	if err := os.MkdirAll(dstMedia, 0o755); err != nil {
		t.Fatalf("mkdir dst media: %v", err)
	}

	mustWriteFile(t, filepath.Join(srcSystem, "Imgs", "Game A.png"), "art")
	mustWriteFile(t, filepath.Join(srcSystem, "Imgs", "Game B.jpg"), "art")

	copied, duplicates, err := copySystemArtwork("FC", srcSystem, dstMedia, "", globalFlags{})
	if err != nil {
		t.Fatalf("copySystemArtwork: %v", err)
	}
	if copied != 1 {
		t.Fatalf("expected 1 copied art file, got %d", copied)
	}
	if duplicates != 0 {
		t.Fatalf("expected 0 duplicates, got %d", duplicates)
	}

	assertFileExists(t, filepath.Join(dstMedia, "Game A.png"))
	assertFileMissing(t, filepath.Join(dstMedia, "Game B.jpg"))
}

func TestCopySystemArtworkMDRoutes32XArtTo32XMedia(t *testing.T) {
	t.Parallel()

	srcSystem := t.TempDir()
	romsRoot := t.TempDir()
	mdMedia := filepath.Join(romsRoot, nextUISystemFolderName("MD"), ".media")
	md32xMedia := filepath.Join(romsRoot, md32XSystemFolderName, ".media")

	if err := os.MkdirAll(mdMedia, 0o755); err != nil {
		t.Fatalf("mkdir md media: %v", err)
	}
	if err := os.MkdirAll(md32xMedia, 0o755); err != nil {
		t.Fatalf("mkdir md32x media: %v", err)
	}

	mustWriteFile(t, filepath.Join(romsRoot, md32XSystemFolderName, "Doom.zip"), "rom")
	mustWriteFile(t, filepath.Join(romsRoot, nextUISystemFolderName("MD"), "Sonic.zip"), "rom")
	mustWriteFile(t, filepath.Join(srcSystem, "Imgs", "Doom.png"), "art")
	mustWriteFile(t, filepath.Join(srcSystem, "Imgs", "Sonic.png"), "art")

	copied, duplicates, err := copySystemArtwork("MD", srcSystem, mdMedia, md32xMedia, globalFlags{})
	if err != nil {
		t.Fatalf("copySystemArtwork md split: %v", err)
	}
	if copied != 2 || duplicates != 0 {
		t.Fatalf("unexpected counts copied=%d duplicates=%d", copied, duplicates)
	}

	assertFileExists(t, filepath.Join(md32xMedia, "Doom.png"))
	assertFileExists(t, filepath.Join(mdMedia, "Sonic.png"))
	assertFileMissing(t, filepath.Join(mdMedia, "Doom.png"))
}

func TestConvertDoneSet3ToNextUICreatesRomsBiosAndMedia(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	romsSrc := filepath.Join(source, "Roms")
	biosSrc := filepath.Join(source, "BIOS")
	destination := filepath.Join(t.TempDir(), "export")

	mustWriteFile(t, filepath.Join(romsSrc, "FC", "All but the Best (NES)", "Game 1.nes"), "rom1")
	mustWriteFile(t, filepath.Join(romsSrc, "FC", "Game 2.nes"), "rom2")
	mustWriteFile(t, filepath.Join(romsSrc, "FC", "Imgs", "Game 1.png"), "art1")
	mustWriteFile(t, filepath.Join(biosSrc, "PS", "scph5501.bin"), "bios")

	stats, err := convertDoneSet3ToNextUI(romsSrc, biosSrc, destination, false, false, globalFlags{})
	if err != nil {
		t.Fatalf("convertDoneSet3ToNextUI: %v", err)
	}
	if stats.Systems != 1 {
		t.Fatalf("expected 1 system, got %d", stats.Systems)
	}
	if stats.ROMSCopied != 2 {
		t.Fatalf("expected 2 ROMs copied, got %d", stats.ROMSCopied)
	}
	if stats.ArtCopied != 1 {
		t.Fatalf("expected 1 art copied, got %d", stats.ArtCopied)
	}
	if stats.BIOSCopied != 1 {
		t.Fatalf("expected 1 BIOS copied, got %d", stats.BIOSCopied)
	}

	romSystem := filepath.Join(destination, "Roms", nextUISystemFolderName("FC"))
	assertFileExists(t, filepath.Join(romSystem, "Game 1.nes"))
	assertFileExists(t, filepath.Join(romSystem, "Game 2.nes"))
	assertFileExists(t, filepath.Join(romSystem, ".media", "Game 1.png"))
	assertFileExists(t, filepath.Join(destination, "Bios", "PS", "scph5501.bin"))
}

func TestRunCuratedConvertValidatesRequiredFlags(t *testing.T) {
	t.Parallel()

	err := runCuratedConvert(globalFlags{}, []string{"--set", "done-set-3", "--target", "nextui"})
	if err == nil {
		t.Fatal("expected missing source/destination error")
	}
	if !strings.Contains(err.Error(), "requires --source") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContainsBoundedKeyword(t *testing.T) {
	t.Parallel()

	if !containsBoundedKeyword("sonic the hedgehog", "sonic") {
		t.Fatal("expected bounded keyword match for sonic title")
	}
	if containsBoundedKeyword("aerobiz supersonic", "sonic") {
		t.Fatal("did not expect false-positive match inside supersonic")
	}
	if !containsBoundedKeyword("nba jam", "nba") {
		t.Fatal("expected bounded keyword match for nba")
	}
	if containsBoundedKeyword("cobna game", "nba") {
		t.Fatal("did not expect substring match inside alphanumeric token")
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parents for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected file %s to be absent", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("expected not-exist for %s, got: %v", path, err)
	}
}

func TestStripROMDisplayTags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		// Extension removal only.
		{"Metroid.nes", "Metroid"},
		{"Metroid.zip", "Metroid"},

		// Region tag removal.
		{"Metroid (USA).nes", "Metroid"},
		{"Super Mario World (USA).sfc", "Super Mario World"},
		{"Sonic the Hedgehog (Europe).zip", "Sonic the Hedgehog"},
		{"Street Fighter II (Japan).sfc", "Street Fighter II"},
		{"Final Fight (World).zip", "Final Fight"},

		// Revision tags.
		{"Mega Man 2 (USA) (Rev A).nes", "Mega Man 2"},
		{"Mega Man 3 (USA) (Rev 1).nes", "Mega Man 3"},
		{"Mega Man X (USA) (v1.1).sfc", "Mega Man X"},

		// Special release tags.
		{"Castlevania (USA) (Beta).nes", "Castlevania"},
		{"Contra (USA) (Proto).nes", "Contra"},
		{"Demo Game (USA) (Demo).nes", "Demo Game"},

		// GoodTools bracket flags.
		{"Sonic the Hedgehog [!].zip", "Sonic the Hedgehog"},
		{"Bad Rom [b].zip", "Bad Rom"},
		{"Alternate [a1].zip", "Alternate"},

		// Combined tags (region + goodtools).
		{"Legend of Zelda, The (USA) [!].nes", "The Legend of Zelda"},

		// Article reattachment.
		{"Legend of Zelda, The (USA).nes", "The Legend of Zelda"},
		{"Hero, A (USA).nes", "A Hero"},
		{"Oddity, An (USA).nes", "An Oddity"},

		// Multi-language parens.
		{"Game (En,Fr,De).zip", "Game"},

		// No tags — should be unchanged besides extension.
		{"Pokemon Red.gb", "Pokemon Red"},

		// Edge case: name that ends with "), The" pattern inside parens (should not trigger article logic).
		{"Pokemon Pinball (USA).gbc", "Pokemon Pinball"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got := stripROMDisplayTags(tc.in)
			if got != tc.want {
				t.Errorf("stripROMDisplayTags(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSystemKeyFromFolderName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		{"06) Nintendo Entertainment System (FC)", "FC"},
		{"16) Game Boy Advance (GBA)", "GBA"},
		{"27) MS-DOS (DOS)", "DOS"},
		{"01) Arcade (ARCADE)", "ARCADE"},
		{"No parens", ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got := systemKeyFromFolderName(tc.in)
			if got != tc.want {
				t.Errorf("systemKeyFromFolderName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestGenerateSystemMapTxt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	g := globalFlags{}

	mustWriteFile(t, filepath.Join(dir, "Metroid (USA).nes"), "rom")
	mustWriteFile(t, filepath.Join(dir, "Mega Man 2 (USA).nes"), "rom")
	mustWriteFile(t, filepath.Join(dir, "Legend of Zelda, The (USA) [!].nes"), "rom")
	mustWriteFile(t, filepath.Join(dir, "Castlevania (USA).nes"), "rom")
	// These should be excluded from map.txt.
	mustWriteFile(t, filepath.Join(dir, "map.txt"), "old")
	mustWriteFile(t, filepath.Join(dir, "thumbs.db"), "junk")

	files, collisions, err := generateSystemMapTxt(dir, g)
	if err != nil {
		t.Fatalf("generateSystemMapTxt: %v", err)
	}
	if files != 1 {
		t.Errorf("expected 1 map file written, got %d", files)
	}
	if collisions != 0 {
		t.Errorf("expected 0 collisions, got %d", collisions)
	}

	mapPath := filepath.Join(dir, "map.txt")
	assertFileExists(t, mapPath)

	content, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatalf("reading map.txt: %v", err)
	}
	got := string(content)

	// Each ROM should appear as "filename|display" — check key lines.
	for _, line := range []string{
		"Metroid (USA).nes|Metroid",
		"Mega Man 2 (USA).nes|Mega Man 2",
		"Legend of Zelda, The (USA) [!].nes|The Legend of Zelda",
		"Castlevania (USA).nes|Castlevania",
	} {
		if !strings.Contains(got, line) {
			t.Errorf("map.txt missing expected line %q\nfull content:\n%s", line, got)
		}
	}

	// thumbs.db and the old map.txt must not appear as entries.
	for _, bad := range []string{"thumbs.db|", "map.txt|"} {
		if strings.Contains(got, bad) {
			t.Errorf("map.txt should not contain %q\nfull content:\n%s", bad, got)
		}
	}
}

func TestGenerateSystemMapTxtCollisions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	g := globalFlags{}

	// Two ROMs that strip to the same display name.
	mustWriteFile(t, filepath.Join(dir, "Game (USA).zip"), "rom")
	mustWriteFile(t, filepath.Join(dir, "Game (Europe).zip"), "rom")
	mustWriteFile(t, filepath.Join(dir, "Unique Title (USA).zip"), "rom")

	_, collisions, err := generateSystemMapTxt(dir, g)
	if err != nil {
		t.Fatalf("generateSystemMapTxt: %v", err)
	}
	if collisions != 1 {
		t.Errorf("expected 1 collision, got %d", collisions)
	}
}

func TestGenerateSystemMapTxtDryRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	g := globalFlags{dryRun: true}

	mustWriteFile(t, filepath.Join(dir, "Metroid (USA).nes"), "rom")

	files, _, err := generateSystemMapTxt(dir, g)
	if err != nil {
		t.Fatalf("generateSystemMapTxt dry-run: %v", err)
	}
	if files != 1 {
		t.Errorf("dry-run should still report 1 file, got %d", files)
	}

	// No actual map.txt should be written on dry-run.
	assertFileMissing(t, filepath.Join(dir, "map.txt"))
}

func TestGenerateAllSystemMapTxtSkipsTreeMode(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	g := globalFlags{}

	// Flat system: should get map.txt.
	flatDir := filepath.Join(root, "16) Game Boy Advance (GBA)")
	mustWriteFile(t, filepath.Join(flatDir, "Pokemon Red (USA).gba"), "rom")

	// Tree-mode system: must be skipped.
	dosDir := filepath.Join(root, "27) MS-DOS (DOS)")
	mustWriteFile(t, filepath.Join(dosDir, "SomeGame", "GAME.EXE"), "exe")

	files, _, err := generateAllSystemMapTxt(root, g)
	if err != nil {
		t.Fatalf("generateAllSystemMapTxt: %v", err)
	}
	if files != 1 {
		t.Errorf("expected exactly 1 map.txt (GBA only), got %d", files)
	}

	assertFileExists(t, filepath.Join(flatDir, "map.txt"))
	assertFileMissing(t, filepath.Join(dosDir, "map.txt"))
}
