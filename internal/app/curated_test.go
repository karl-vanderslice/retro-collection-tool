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

	copied, duplicates, converted, err := copySystemROMs("FC", src, dst, systemModeFlatten, globalFlags{})
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

	copied, duplicates, converted, err := copySystemROMs("MD", src, dst, systemModeFlatten, globalFlags{})
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

	copied, duplicates, converted, err := copySystemROMs("MD", src, mdDst, systemModeFlatten, globalFlags{})
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

	copied, duplicates, converted, err := copySystemROMs("SCUMMVM", src, dst, systemModeTree, globalFlags{})
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

	copied, duplicates, converted, err := copySystemROMs("PS", src, dst, systemModeFlatten, globalFlags{})
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

	copied, duplicates, converted, err := copySystemROMs("ARCADE", src, dst, systemModeFlatten, globalFlags{})
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

	copied, duplicates, converted, err := copySystemROMs("FC", src, dst, systemModeFlatten, globalFlags{dryRun: true})
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

	copied, duplicates, converted, err := copySystemROMs("ZXS", src, dst, systemModeFlatten, globalFlags{})
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

	copied, duplicates, converted, err := copySystemROMs("FC", src, dst, systemModeFlatten, globalFlags{})
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

	stats, err := convertDoneSet3ToNextUI(romsSrc, biosSrc, destination, globalFlags{})
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
