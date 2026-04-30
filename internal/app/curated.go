package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/karl-vanderslice/retro-collection-tool/internal/fsutil"
)

// romGoodToolsTagRe matches GoodTools bracket flags like [!], [b], [a1] at the end of a name.
var romGoodToolsTagRe = regexp.MustCompile(`\s*\[[^\]]{0,15}\]\s*$`)

// romRegionRevTagRe matches common region and revision parentheticals at the end of a name,
// e.g. (USA), (Europe), (Japan), (Rev A), (v1.1), (Beta), (Proto), (Demo), (En,Fr,De).
var romRegionRevTagRe = regexp.MustCompile(`(?i)\s*\((?:USA|U\.S\.A\.|Europe|EUR|Eu|Japan|JPN|Jpn|World|UK|Australia|Aus|Brazil|BRA|Korea|KOR|China|Asia|Canada|France|Germany|Spain|Italy|Sweden|Netherlands|Denmark|Norway|Finland|En(?:,[A-Za-z,]+)?|Fr|De|Ja|Es|It|Pt|Nl|Ko|Zh|Sv|No|Da|Rev\.?\s*[A-Z0-9]+|v\d+\.\d+[^)]*|Beta[^)]*|Proto(?:type)?[^)]*|Demo[^)]*|Sample[^)]*|Unl|Unlicensed|Pirate|Kiosk|Virtual Console)\)\s*$`)

var excludedROMFolders = map[string]bool{
	"imgs":                true,
	"manuals":             true,
	"translations":        true,
	"unlicensed homebrew": true,
	"hacks":               true,
	"__macosx":            true,
}

var excludedROMFilenames = map[string]bool{
	".ds_store":   true,
	"thumbs.db":   true,
	"desktop.ini": true,
	"ehthumbs.db": true,
}

const md32XSystemFolderName = "10) Sega 32X (32X)"

const md32XSystemKey = "X32"

type systemMenuSpec struct {
	Order int
	Title string
	Tag   string
}

var systemMenuSpecs = map[string]systemMenuSpec{
	"ARCADE":              {Order: 0, Title: "Arcade", Tag: "ARCADE"},
	"ATARI":               {Order: 1, Title: "Atari 2600", Tag: "ATARI"},
	"FIFTYTWOHUNDRED":     {Order: 2, Title: "Atari 5200", Tag: "FIFTYTWOHUNDRED"},
	"SEVENTYEIGHTHUNDRED": {Order: 3, Title: "Atari 7800", Tag: "SEVENTYEIGHTHUNDRED"},
	"COLECO":              {Order: 4, Title: "ColecoVision", Tag: "COLECO"},
	"VECTREX":             {Order: 5, Title: "Vectrex", Tag: "VECTREX"},
	"FC":                  {Order: 6, Title: "Nintendo Entertainment System", Tag: "FC"},
	"FDS":                 {Order: 7, Title: "Famicom Disk System", Tag: "FDS"},
	"MS":                  {Order: 8, Title: "Sega Master System", Tag: "MS"},
	"MD":                  {Order: 9, Title: "Sega Genesis", Tag: "MD"},
	md32XSystemKey:        {Order: 10, Title: "Sega 32X", Tag: "32X"},
	"SEGACD":              {Order: 11, Title: "Sega CD", Tag: "SEGACD"},
	"SFC":                 {Order: 12, Title: "Super Nintendo Entertainment System", Tag: "SFC"},
	"SATELLAVIEW":         {Order: 13, Title: "Satellaview", Tag: "SATELLAVIEW"},
	"GB":                  {Order: 14, Title: "Game Boy", Tag: "GB"},
	"GBC":                 {Order: 15, Title: "Game Boy Color", Tag: "GBC"},
	"GBA":                 {Order: 16, Title: "Game Boy Advance", Tag: "GBA"},
	"GG":                  {Order: 17, Title: "Game Gear", Tag: "GG"},
	"GW":                  {Order: 18, Title: "Game & Watch", Tag: "GW"},
	"LYNX":                {Order: 19, Title: "Atari Lynx", Tag: "LYNX"},
	"PCE":                 {Order: 20, Title: "TurboGrafx-16", Tag: "PCE"},
	"PCECD":               {Order: 21, Title: "TurboGrafx-CD", Tag: "PCECD"},
	"NEOGEO":              {Order: 22, Title: "Neo Geo", Tag: "NEOGEO"},
	"PS":                  {Order: 23, Title: "Sony PlayStation", Tag: "PS"},
	"NDS":                 {Order: 24, Title: "Nintendo DS", Tag: "NDS"},
	"COMMODORE":           {Order: 25, Title: "Commodore 64", Tag: "COMMODORE"},
	"ZXS":                 {Order: 26, Title: "ZX Spectrum", Tag: "ZXS"},
	"DOS":                 {Order: 27, Title: "MS-DOS", Tag: "DOS"},
	"SCUMMVM":             {Order: 28, Title: "ScummVM", Tag: "SCUMMVM"},
	"PORTS":               {Order: 29, Title: "Ports", Tag: "PORTS"},
	"PICO":                {Order: 30, Title: "PICO-8", Tag: "PICO"},
}

type collectionSpec struct {
	Name     string
	Keywords []string
}

var collectionSpecs = []collectionSpec{
	{Name: "Mario Series", Keywords: []string{"super mario", "dr. mario", "paper mario", "mario kart", "mario golf", "mario tennis", "mario party", "wario", "yoshi"}},
	{Name: "The Legend of Zelda Series", Keywords: []string{"zelda", "link to the past", "ocarina", "majora", "hyrule"}},
	{Name: "Metroid Series", Keywords: []string{"metroid"}},
	{Name: "Pokemon Series", Keywords: []string{"pokemon", "pok\u00e9mon"}},
	{Name: "Donkey Kong Series", Keywords: []string{"donkey kong", "diddy kong"}},
	{Name: "Kirby Series", Keywords: []string{"kirby"}},
	{Name: "F-Zero Series", Keywords: []string{"f-zero", "f zero"}},
	{Name: "Fire Emblem Series", Keywords: []string{"fire emblem"}},
	{Name: "EarthBound and Mother Series", Keywords: []string{"earthbound", "mother"}},
	{Name: "Star Fox Series", Keywords: []string{"star fox", "starfox"}},
	{Name: "Final Fantasy Series", Keywords: []string{"final fantasy", "ff tactics", "final fantasy legend", "mystic quest"}},
	{Name: "Dragon Quest Series", Keywords: []string{"dragon quest", "dragon warrior"}},
	{Name: "Chrono Series", Keywords: []string{"chrono trigger", "chrono cross"}},
	{Name: "Mana Series", Keywords: []string{"secret of mana", "trials of mana", "seiken densetsu", "legend of mana"}},
	{Name: "Castlevania Series", Keywords: []string{"castlevania", "akumajou dracula", "kid dracula"}},
	{Name: "Mega Man Series", Keywords: []string{"mega man", "megaman", "rockman"}},
	{Name: "Contra Series", Keywords: []string{"contra", "probotector"}},
	{Name: "Metal Gear Series", Keywords: []string{"metal gear"}},
	{Name: "Resident Evil Series", Keywords: []string{"resident evil", "biohazard"}},
	{Name: "Silent Hill Series", Keywords: []string{"silent hill"}},
	{Name: "Tomb Raider Series", Keywords: []string{"tomb raider", "lara croft"}},
	{Name: "Sonic the Hedgehog Series", Keywords: []string{"sonic", "knuckles", "tails", "dr. robotnik", "eggman"}},
	{Name: "Golden Axe Series", Keywords: []string{"golden axe"}},
	{Name: "Streets of Rage Series", Keywords: []string{"streets of rage", "bare knuckle"}},
	{Name: "Final Fight Series", Keywords: []string{"final fight", "haggar", "mad gear"}},
	{Name: "Phantasy Star Series", Keywords: []string{"phantasy star"}},
	{Name: "Street Fighter Series", Keywords: []string{"street fighter", "sfii", "sf zero", "alpha 3"}},
	{Name: "Mortal Kombat Series", Keywords: []string{"mortal kombat", "mk trilogy", "mk mythologies"}},
	{Name: "Pinball Games", Keywords: []string{"pinball", "alien crush", "devil's crush", "devil's crush", "dragon's fury", "sonic spinball"}},
	{Name: "Card Games", Keywords: []string{"trading card", "card gb", "card game", "yu-gi-oh", "yugioh", "mahjong", "solitaire", "poker", "blackjack", "duelist", "duel academy"}},
	{Name: "Beat \u2018Em Ups", Keywords: []string{"double dragon", "river city ransom", "kunio", "captain commando", "knights of the round", "battletoads", "turtles in time", "cadillacs and dinosaurs", "armored warriors", "avenger"}},
	{Name: "King of Fighters Series", Keywords: []string{"king of fighters", "kof"}},
	{Name: "Samurai Shodown Series", Keywords: []string{"samurai shodown", "samurai spirits"}},
	{Name: "Metal Slug Series", Keywords: []string{"metal slug"}},
	{Name: "Pac-Man Series", Keywords: []string{"pac-man", "pac man", "pacmania"}},
	{Name: "Bomberman Series", Keywords: []string{"bomberman"}},
	{Name: "Gradius Series", Keywords: []string{"gradius", "nemesis"}},
	{Name: "R-Type Series", Keywords: []string{"r-type", "rtype"}},
	{Name: "Galaga and Galaxian Series", Keywords: []string{"galaga", "galaxian"}},
	{Name: "Batman Games", Keywords: []string{"batman", "dark knight"}},
	{Name: "Marvel Games", Keywords: []string{"marvel", "x-men", "x men", "spider-man", "spiderman", "captain america", "iron man", "hulk", "punisher", "avengers"}},
	{Name: "DC Comics Games", Keywords: []string{"superman", "justice league", "flash", "green lantern", "wonder woman", "aquaman", "teen titans"}},
	{Name: "Disney Games", Keywords: []string{"disney", "mickey", "minnie", "donald", "goofy", "aladdin", "lion king", "little mermaid", "ducktales", "chip 'n dale", "chip n dale"}},
	{Name: "Star Wars Games", Keywords: []string{"star wars"}},
	{Name: "Indiana Jones Games", Keywords: []string{"indiana jones"}},
	{Name: "Jurassic Park Games", Keywords: []string{"jurassic park", "jurassic"}},
	{Name: "Teenage Mutant Ninja Turtles Games", Keywords: []string{"teenage mutant ninja turtles", "tmnt", "turtles in time"}},
	{Name: "Madden NFL Games", Keywords: []string{"madden", "john madden football"}},
	{Name: "NHL Hockey Games", Keywords: []string{"nhl", "nhlpa", "wayne gretzky hockey"}},
	{Name: "NBA Basketball Games", Keywords: []string{"nba", "nba jam", "nba live"}},
	{Name: "FIFA and Soccer Games", Keywords: []string{"fifa", "world cup", "iss pro", "pro evolution soccer", "international superstar soccer"}},
	{Name: "Tony Hawk and Skateboarding Games", Keywords: []string{"tony hawk", "thps", "skate or die"}},
}

var excludedROMFileExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".pdf":  true,
	".txt":  true,
	".nfo":  true,
	".md":   true,
	".xml":  true,
	".json": true,
	".db":   true,
	".cfg":  true,
}

const (
	systemModeFlatten = "flatten"
	systemModeTree    = "tree"
)

var systemModes = map[string]string{
	"dos":     systemModeTree,
	"scummvm": systemModeTree,
	"ports":   systemModeTree,
}

var zipUniformExtensions = map[string]bool{
	".gba": true,
	".gb":  true,
	".gbc": true,
	".sfc": true,
	".smc": true,
	".md":  true,
	".sms": true,
	".gg":  true,
	".pce": true,
	".sgx": true,
	".ws":  true,
	".wsc": true,
}

type curatedConvertStats struct {
	Systems        int
	ROMSCopied     int
	ROMDuplicates  int
	CorruptROMs    int
	ArtCopied      int
	ArtDuplicates  int
	BIOSCopied     int
	SevenZToZip    int
	RawToZip       int
	RezippedROMs   int
	Collections    int
	CollectionROMs int
	ArcadeMap      bool
	MapTxtFiles    int
	MapCollisions  int
}

func runCurated(g globalFlags, args []string) error {
	if len(args) == 0 {
		return errors.New("curated requires subcommand: convert")
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "convert":
		return runCuratedConvert(g, args[1:])
	default:
		return fmt.Errorf("unknown curated subcommand: %s", args[0])
	}
}

func runCuratedConvert(g globalFlags, args []string) error {
	fs := flag.NewFlagSet("curated convert", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	setName := fs.String("set", "done-set-3", "curated set name (currently supported: done-set-3)")
	target := fs.String("target", "nextui", "target firmware layout (currently supported: nextui)")
	source := fs.String("source", "", "source root path containing Done Set 3 folders (Roms, BIOS)")
	destination := fs.String("destination", "", "destination root path for export")
	full := fs.Bool("full", false, "wipe and rebuild the full destination before conversion (default: incremental, skip existing files)")
	permanent := fs.Bool("permanent", false, "no hard links; recompress all ROMs to maximum zip compression (for archival storage)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ensureNoPositionalArgs("curated convert", fs.Args()); err != nil {
		return err
	}

	if strings.ToLower(strings.TrimSpace(*setName)) != "done-set-3" {
		return fmt.Errorf("unsupported --set %q (supported: done-set-3)", *setName)
	}
	if strings.ToLower(strings.TrimSpace(*target)) != "nextui" {
		return fmt.Errorf("unsupported --target %q (supported: nextui)", *target)
	}
	if strings.TrimSpace(*source) == "" {
		return errors.New("curated convert requires --source")
	}
	if strings.TrimSpace(*destination) == "" {
		return errors.New("curated convert requires --destination")
	}

	srcRoot := filepath.Clean(*source)
	dstRoot := filepath.Clean(*destination)

	romsSrc := filepath.Join(srcRoot, "Roms")
	biosSrc := filepath.Join(srcRoot, "BIOS")

	if err := requireDir(romsSrc); err != nil {
		return err
	}
	if err := requireDir(biosSrc); err != nil {
		return err
	}

	stats, err := convertDoneSet3ToNextUI(romsSrc, biosSrc, dstRoot, *full, *permanent, g)
	if err != nil {
		return err
	}

	emitInfo(g, "curated", "convert", "summary", outputFields{
		"set":               "done-set-3",
		"target":            "nextui",
		"source":            srcRoot,
		"destination":       dstRoot,
		"systems":           stats.Systems,
		"roms_copied":       stats.ROMSCopied,
		"rom_duplicates":    stats.ROMDuplicates,
		"art_copied":        stats.ArtCopied,
		"art_duplicates":    stats.ArtDuplicates,
		"bios_files_copied": stats.BIOSCopied,
		"seven_z_to_zip":    stats.SevenZToZip,
		"raw_to_zip":        stats.RawToZip,
		"rezipped_roms":     stats.RezippedROMs,
		"corrupt_roms":      stats.CorruptROMs,
		"collections":       stats.Collections,
		"collection_roms":   stats.CollectionROMs,
		"arcade_map":        stats.ArcadeMap,
		"map_txt_files":     stats.MapTxtFiles,
		"map_collisions":    stats.MapCollisions,
		"mode": func() string {
			base := "incremental"
			if *full {
				base = "full"
			}
			if *permanent {
				return base + "+permanent"
			}
			return base
		}(),
		"dry_run": g.dryRun,
	})

	return nil
}

func convertDoneSet3ToNextUI(romsSrc, biosSrc, destination string, full, permanent bool, g globalFlags) (curatedConvertStats, error) {
	stats := curatedConvertStats{}

	romsDstRoot := filepath.Join(destination, "Roms")
	biosDstRoot := filepath.Join(destination, "Bios")

	if g.dryRun {
		if full {
			emitInfo(g, "curated", "convert", "dry-run clean output roots", outputFields{"roms": romsDstRoot, "bios": biosDstRoot})
		}
		emitInfo(g, "curated", "convert", "dry-run prepare output roots", outputFields{"roms": romsDstRoot, "bios": biosDstRoot})
	} else if full {
		if err := fsutil.RemoveIfExists(romsDstRoot); err != nil {
			return stats, fmt.Errorf("clean ROM output root %s: %w", romsDstRoot, err)
		}
		if err := fsutil.RemoveIfExists(biosDstRoot); err != nil {
			return stats, fmt.Errorf("clean BIOS output root %s: %w", biosDstRoot, err)
		}
		if err := fsutil.EnsureDir(romsDstRoot); err != nil {
			return stats, err
		}
		if err := fsutil.EnsureDir(biosDstRoot); err != nil {
			return stats, err
		}
	} else {
		if err := fsutil.EnsureDir(romsDstRoot); err != nil {
			return stats, err
		}
		if err := fsutil.EnsureDir(biosDstRoot); err != nil {
			return stats, err
		}
	}

	systems, err := os.ReadDir(romsSrc)
	if err != nil {
		return stats, fmt.Errorf("read rom systems: %w", err)
	}

	processedSystemDirs := map[string]bool{}

	for _, system := range systems {
		if !system.IsDir() {
			continue
		}
		systemName := strings.TrimSpace(system.Name())
		if systemName == "" || strings.HasPrefix(systemName, ".") {
			continue
		}

		stats.Systems++
		srcSystemDir := filepath.Join(romsSrc, systemName)
		dstKey := canonicalDestinationSystemKey(systemName)
		dstSystemDir := filepath.Join(romsDstRoot, nextUISystemFolderName(dstKey))
		dstMediaDir := filepath.Join(dstSystemDir, ".media")
		dst32XSystemDir := md32XDestinationDir(romsDstRoot)
		dst32XMediaDir := filepath.Join(dst32XSystemDir, ".media")
		systemMode := resolveSystemMode(systemName)

		if g.dryRun {
			emitInfo(g, "curated", "convert", "dry-run system mapping", outputFields{"system": systemName, "mode": systemMode, "from": srcSystemDir, "to": dstSystemDir})
			if dst32XSystemDir != "" {
				emitInfo(g, "curated", "convert", "dry-run system mapping", outputFields{"system": systemName, "mode": systemMode, "from": filepath.Join(srcSystemDir, "32X Games (Genesis)"), "to": dst32XSystemDir})
			}
		} else {
			if !processedSystemDirs[dstSystemDir] {
				if err := fsutil.EnsureDir(dstSystemDir); err != nil {
					return stats, err
				}
				if err := fsutil.EnsureDir(dstMediaDir); err != nil {
					return stats, err
				}
				processedSystemDirs[dstSystemDir] = true
			}
			if dst32XSystemDir != "" {
				if !processedSystemDirs[dst32XSystemDir] {
					if err := fsutil.EnsureDir(dst32XSystemDir); err != nil {
						return stats, err
					}
					if err := fsutil.EnsureDir(dst32XMediaDir); err != nil {
						return stats, err
					}
					processedSystemDirs[dst32XSystemDir] = true
				}
			}
		}

		rc, rd, zc, rz2, sk, err := copySystemROMs(systemName, srcSystemDir, dstSystemDir, systemMode, permanent, g)
		if err != nil {
			return stats, err
		}
		stats.ROMSCopied += rc
		stats.ROMDuplicates += rd
		stats.SevenZToZip += zc
		stats.RezippedROMs += rz2
		stats.CorruptROMs += sk

		if !permanent {
			rz, err := normalizeSystemZipUniformity(dstSystemDir, systemMode, g)
			if err != nil {
				return stats, err
			}
			stats.RawToZip += rz
			if dst32XSystemDir != "" {
				rz32x, err := normalizeSystemZipUniformity(dst32XSystemDir, systemModeFlatten, g)
				if err != nil {
					return stats, err
				}
				stats.RawToZip += rz32x
			}
		}

		ac, ad, err := copySystemArtwork(systemName, srcSystemDir, dstMediaDir, dst32XMediaDir, g)
		if err != nil {
			return stats, err
		}
		stats.ArtCopied += ac
		stats.ArtDuplicates += ad
	}

	hasArcadeMap, err := ensureArcadeMap(romsSrc, romsDstRoot, g)
	if err != nil {
		return stats, err
	}
	stats.ArcadeMap = hasArcadeMap

	mapFiles, mapCollisions, err := generateAllSystemMapTxt(romsDstRoot, g)
	if err != nil {
		return stats, err
	}
	stats.MapTxtFiles = mapFiles
	stats.MapCollisions = mapCollisions

	biosCopied, err := copyDirectoryFiles(biosSrc, biosDstRoot, g)
	if err != nil {
		return stats, err
	}
	stats.BIOSCopied = biosCopied

	collections, entries, err := writeCollections(destination, g)
	if err != nil {
		return stats, err
	}
	stats.Collections = collections
	stats.CollectionROMs = entries

	return stats, nil
}

func copySystemROMs(systemName, srcSystemDir, dstSystemDir, mode string, permanent bool, g globalFlags) (copied int, duplicates int, sevenZToZip int, rezipped int, skipped int, err error) {
	if strings.TrimSpace(mode) == "" {
		mode = systemModeFlatten
	}

	err = filepath.WalkDir(srcSystemDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, relErr := filepath.Rel(srcSystemDir, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}

		parts := splitPath(rel)
		if shouldSkipROMPath(parts, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if excludedROMFilenames[strings.ToLower(strings.TrimSpace(d.Name()))] {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if strings.EqualFold(mode, systemModeFlatten) && excludedROMFileExtensions[ext] {
			return nil
		}

		dst := targetPathForROM(systemName, mode, dstSystemDir, rel, d.Name())
		if ext == ".7z" || (permanent && zipUniformExtensions[ext]) {
			dst = strings.TrimSuffix(dst, filepath.Ext(dst)) + ".zip"
		}

		exists, existsErr := fileExistsAt(dst)
		if existsErr != nil {
			return existsErr
		}
		if exists {
			duplicates++
			return nil
		}

		if g.dryRun {
			switch {
			case ext == ".7z":
				emitVerbose(g, "curated", "convert", "dry-run 7z->zip", outputFields{"system": systemName, "from": path, "to": dst})
				sevenZToZip++
				copied++
			case permanent && ext == ".zip":
				emitVerbose(g, "curated", "convert", "dry-run rezip-max", outputFields{"system": systemName, "from": path, "to": dst})
				rezipped++
				copied++
			case permanent && zipUniformExtensions[ext]:
				emitVerbose(g, "curated", "convert", "dry-run raw->zip-max", outputFields{"system": systemName, "from": path, "to": dst})
				rezipped++
				copied++
			default:
				emitVerbose(g, "curated", "convert", "dry-run ROM copy", outputFields{"system": systemName, "from": path, "to": dst})
				copied++
			}
			return nil
		}

		switch {
		case ext == ".7z":
			level := 0
			if permanent {
				level = 9
			}
			if err := convert7zToZip(path, dst, level); err != nil {
				emitInfo(g, "curated", "convert", "skipping corrupt archive", outputFields{"system": systemName, "path": path, "error": err.Error()})
				skipped++
				return nil
			}
			sevenZToZip++
			copied++
		case permanent && ext == ".zip":
			if err := convert7zToZip(path, dst, 9); err != nil {
				emitInfo(g, "curated", "convert", "skipping corrupt archive", outputFields{"system": systemName, "path": path, "error": err.Error()})
				skipped++
				return nil
			}
			rezipped++
			copied++
		case permanent && zipUniformExtensions[ext]:
			if err := packRawROMToZip(path, dst, 9); err != nil {
				return err
			}
			rezipped++
			copied++
		default:
			if permanent {
				if err := fsutil.CopyFile(path, dst); err != nil {
					return err
				}
			} else {
				if err := fsutil.LinkOrCopy(path, dst); err != nil {
					return err
				}
			}
			copied++
		}
		return nil
	})

	return copied, duplicates, sevenZToZip, rezipped, skipped, err
}

func targetPathForROM(systemName, mode, dstSystemDir, rel, basename string) string {
	if shouldPreserveArcadeCHDPath(systemName, rel) {
		return filepath.Join(dstSystemDir, rel)
	}
	if isMD32XContent(systemName, rel) {
		return filepath.Join(md32XDestinationDirFromMD(dstSystemDir), basename)
	}
	if strings.EqualFold(mode, systemModeTree) {
		return filepath.Join(dstSystemDir, rel)
	}
	if strings.HasPrefix(rel, ".hidden") || strings.Contains(rel, string(os.PathSeparator)+".hidden"+string(os.PathSeparator)) {
		return filepath.Join(dstSystemDir, rel)
	}
	return filepath.Join(dstSystemDir, basename)
}

func shouldPreserveArcadeCHDPath(systemName, rel string) bool {
	if !strings.EqualFold(canonicalDestinationSystemKey(systemName), "ARCADE") {
		return false
	}
	return strings.EqualFold(filepath.Ext(rel), ".chd")
}

func md32XDestinationDir(romsDstRoot string) string {
	return filepath.Join(romsDstRoot, nextUISystemFolderName(md32XSystemKey))
}

func md32XDestinationDirFromMD(mdSystemDir string) string {
	return filepath.Join(filepath.Dir(mdSystemDir), nextUISystemFolderName(md32XSystemKey))
}

func isMD32XContent(systemName, rel string) bool {
	if !strings.EqualFold(strings.TrimSpace(systemName), "MD") {
		return false
	}
	parts := splitPath(rel)
	if len(parts) == 0 {
		return false
	}
	return strings.EqualFold(normalizeFolderCategory(parts[0]), "32x games")
}

func resolveSystemMode(systemName string) string {
	key := strings.ToLower(strings.TrimSpace(systemName))
	if mode, ok := systemModes[key]; ok {
		return mode
	}
	return systemModeFlatten
}

func convert7zToZip(src7zPath, dstZipPath string, level int) error {
	if _, err := exec.LookPath("7z"); err != nil {
		return fmt.Errorf("7z not found in PATH for conversion (%s): %w", src7zPath, err)
	}

	if err := fsutil.EnsureDir(filepath.Dir(dstZipPath)); err != nil {
		return err
	}

	workDir, err := os.MkdirTemp("", "rct-7z-")
	if err != nil {
		return fmt.Errorf("create temp dir for 7z conversion: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(workDir)
	}()

	extract := exec.Command("7z", "x", "-y", "-o"+workDir, src7zPath)
	if out, err := extract.CombinedOutput(); err != nil {
		return fmt.Errorf("extract 7z %s: %w: %s", src7zPath, err, strings.TrimSpace(string(out)))
	}

	entries, err := os.ReadDir(workDir)
	if err != nil {
		return fmt.Errorf("read extracted files for %s: %w", src7zPath, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("7z archive %s extracted no files", src7zPath)
	}

	args := []string{"a", "-tzip", fmt.Sprintf("-mx=%d", level), dstZipPath}
	for _, entry := range entries {
		args = append(args, entry.Name())
	}

	repack := exec.Command("7z", args...)
	repack.Dir = workDir
	if out, err := repack.CombinedOutput(); err != nil {
		return fmt.Errorf("pack zip %s from %s: %w: %s", dstZipPath, src7zPath, err, strings.TrimSpace(string(out)))
	}

	return nil
}

// packRawROMToZip packs a single raw ROM file into a new zip at dstZipPath using the
// given compression level (0=store, 9=max). Unlike zipSingleFile, it reads from an
// arbitrary srcPath and does not delete the source file.
func packRawROMToZip(srcPath, dstZipPath string, level int) error {
	if _, err := exec.LookPath("7z"); err != nil {
		return fmt.Errorf("7z not found in PATH for zip packing (%s): %w", srcPath, err)
	}

	if err := fsutil.EnsureDir(filepath.Dir(dstZipPath)); err != nil {
		return err
	}

	args := []string{"a", "-tzip", fmt.Sprintf("-mx=%d", level), dstZipPath, filepath.Base(srcPath)}
	pack := exec.Command("7z", args...)
	pack.Dir = filepath.Dir(srcPath)
	if out, err := pack.CombinedOutput(); err != nil {
		return fmt.Errorf("zip %s to %s: %w: %s", srcPath, dstZipPath, err, strings.TrimSpace(string(out)))
	}

	return nil
}

func zipSingleFile(srcPath, dstZipPath string) error {
	if _, err := exec.LookPath("7z"); err != nil {
		return fmt.Errorf("7z not found in PATH for zip conversion (%s): %w", srcPath, err)
	}

	if err := fsutil.EnsureDir(filepath.Dir(dstZipPath)); err != nil {
		return err
	}

	baseDir := filepath.Dir(srcPath)
	baseName := filepath.Base(srcPath)

	args := []string{"a", "-tzip", "-mx=0", dstZipPath, baseName}
	pack := exec.Command("7z", args...)
	pack.Dir = baseDir
	if out, err := pack.CombinedOutput(); err != nil {
		return fmt.Errorf("zip %s to %s: %w: %s", srcPath, dstZipPath, err, strings.TrimSpace(string(out)))
	}

	if err := os.Remove(srcPath); err != nil {
		return fmt.Errorf("remove source after zip conversion %s: %w", srcPath, err)
	}

	return nil
}

func normalizeSystemZipUniformity(dstSystemDir, mode string, g globalFlags) (int, error) {
	if !strings.EqualFold(mode, systemModeFlatten) {
		return 0, nil
	}

	entries, err := os.ReadDir(dstSystemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	zipCount := 0
	candidates := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".zip":
			zipCount++
		default:
			if zipUniformExtensions[ext] {
				candidates = append(candidates, entry.Name())
			}
		}
	}

	if zipCount == 0 || len(candidates) == 0 || zipCount <= len(candidates) {
		return 0, nil
	}

	converted := 0
	for _, name := range candidates {
		src := filepath.Join(dstSystemDir, name)
		dst := filepath.Join(dstSystemDir, strings.TrimSuffix(name, filepath.Ext(name))+".zip")

		exists, err := fileExistsAt(dst)
		if err != nil {
			return converted, err
		}
		if exists {
			continue
		}

		if g.dryRun {
			emitVerbose(g, "curated", "convert", "dry-run normalize raw->zip", outputFields{"from": src, "to": dst})
			converted++
			continue
		}

		if err := zipSingleFile(src, dst); err != nil {
			return converted, err
		}
		converted++
	}

	return converted, nil
}

func copySystemArtwork(systemName, srcSystemDir, dstMediaDir, dst32XMediaDir string, g globalFlags) (copied int, duplicates int, err error) {
	imgsDir := filepath.Join(srcSystemDir, "Imgs")
	if statErr := requireDirIfPresent(imgsDir); statErr != nil {
		return 0, 0, statErr
	}
	if _, statErr := os.Stat(imgsDir); os.IsNotExist(statErr) {
		return 0, 0, nil
	}

	md32XROMNames := map[string]bool{}
	if strings.EqualFold(strings.TrimSpace(systemName), "MD") && strings.TrimSpace(dst32XMediaDir) != "" {
		var namesErr error
		md32XROMNames, namesErr = romBasenamesInDir(filepath.Dir(dst32XMediaDir))
		if namesErr != nil {
			return 0, 0, namesErr
		}
	}

	err = filepath.WalkDir(imgsDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		if excludedROMFilenames[strings.ToLower(strings.TrimSpace(d.Name()))] {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".png" {
			return nil
		}

		dst := filepath.Join(dstMediaDir, d.Name())
		if md32XROMNames[strings.TrimSuffix(d.Name(), ext)] {
			dst = filepath.Join(dst32XMediaDir, d.Name())
		}
		exists, existsErr := fileExistsAt(dst)
		if existsErr != nil {
			return existsErr
		}
		if exists {
			duplicates++
			return nil
		}

		if g.dryRun {
			emitVerbose(g, "curated", "convert", "dry-run art copy", outputFields{"from": path, "to": dst})
			copied++
			return nil
		}

		if err := fsutil.CopyFile(path, dst); err != nil {
			return err
		}
		copied++
		return nil
	})

	return copied, duplicates, err
}

func romBasenamesInDir(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, err
	}

	result := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext == "" {
			continue
		}
		result[strings.TrimSuffix(name, ext)] = true
	}
	return result, nil
}

func copyDirectoryFiles(srcRoot, dstRoot string, g globalFlags) (int, error) {
	copied := 0
	err := filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		dst := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			if g.dryRun {
				return nil
			}
			return fsutil.EnsureDir(dst)
		}

		exists, existsErr := fileExistsAt(dst)
		if existsErr != nil {
			return existsErr
		}
		if exists {
			return nil
		}

		if g.dryRun {
			emitVerbose(g, "curated", "convert", "dry-run BIOS copy", outputFields{"from": path, "to": dst})
			copied++
			return nil
		}

		if err := fsutil.CopyFile(path, dst); err != nil {
			return err
		}
		copied++
		return nil
	})

	return copied, err
}

func nextUISystemFolderName(system string) string {
	key := strings.ToUpper(strings.TrimSpace(system))
	if spec, ok := systemMenuSpecs[key]; ok {
		return fmt.Sprintf("%02d) %s (%s)", spec.Order, spec.Title, spec.Tag)
	}
	norm := strings.TrimSpace(system)
	if norm == "" {
		return norm
	}
	if strings.Contains(norm, "(") && strings.Contains(norm, ")") {
		return norm
	}
	return fmt.Sprintf("%s (%s)", norm, norm)
}

func canonicalDestinationSystemKey(systemName string) string {
	key := strings.ToUpper(strings.TrimSpace(systemName))
	switch key {
	case "CPS3", "NEOGEO":
		return "ARCADE"
	default:
		return key
	}
}

func shouldSkipROMPath(pathParts []string, isDir bool) bool {
	if len(pathParts) == 0 {
		return false
	}
	for _, p := range pathParts {
		if strings.HasPrefix(p, ".") && !strings.EqualFold(p, ".hidden") {
			return true
		}
	}
	if len(pathParts) > 0 {
		first := normalizeFolderCategory(pathParts[0])
		if excludedROMFolders[first] {
			return true
		}
	}
	_ = isDir
	return false
}

func normalizeFolderCategory(name string) string {
	norm := strings.ToLower(strings.TrimSpace(name))
	if idx := strings.Index(norm, " ("); idx > 0 {
		norm = strings.TrimSpace(norm[:idx])
	}
	return norm
}

// stripROMDisplayTags removes the file extension plus common trailing region, revision,
// and GoodTools tags from a ROM filename to produce a clean display name.
// Examples:
//
//	Metroid (USA).nes          -> Metroid
//	Super Mario World (USA).sfc -> Super Mario World
//	Sonic the Hedgehog [!].zip -> Sonic the Hedgehog
//	Legend of Zelda, The (USA).zip -> The Legend of Zelda
func stripROMDisplayTags(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.TrimSpace(name)

	// Iteratively strip GoodTools bracket flags: [!], [b], [a1], etc.
	for {
		trimmed := strings.TrimSpace(romGoodToolsTagRe.ReplaceAllString(name, ""))
		if trimmed == name {
			break
		}
		name = trimmed
	}

	// Iteratively strip known region/revision parens: (USA), (Rev A), (v1.1), etc.
	for {
		trimmed := strings.TrimSpace(romRegionRevTagRe.ReplaceAllString(name, ""))
		if trimmed == name {
			break
		}
		name = trimmed
	}

	// Handle ROM sorting convention "Title, The" -> "The Title".
	if strings.HasSuffix(name, ", The") {
		name = "The " + strings.TrimSuffix(name, ", The")
	} else if strings.HasSuffix(name, ", A") {
		name = "A " + strings.TrimSuffix(name, ", A")
	} else if strings.HasSuffix(name, ", An") {
		name = "An " + strings.TrimSuffix(name, ", An")
	}

	return strings.TrimSpace(name)
}

// systemKeyFromFolderName extracts the system key tag from a numbered folder name
// like "06) Nintendo Entertainment System (FC)" -> "FC".
func systemKeyFromFolderName(folderName string) string {
	if !strings.HasSuffix(folderName, ")") {
		return ""
	}
	start := strings.LastIndex(folderName, "(")
	if start < 0 {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(folderName[start+1 : len(folderName)-1]))
}

// generateSystemMapTxt writes a map.txt file into dstSystemDir mapping each ROM
// filename to its clean display name (extension + region/revision tags stripped).
// Returns the number of map files written and the number of display-name collisions found.
func generateSystemMapTxt(dstSystemDir string, g globalFlags) (files int, collisions int, err error) {
	entries, err := os.ReadDir(dstSystemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	type mapEntry struct {
		filename    string
		displayName string
	}

	var mapEntries []mapEntry
	displayCount := make(map[string]int)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if excludedROMFilenames[strings.ToLower(strings.TrimSpace(name))] {
			continue
		}
		if strings.EqualFold(name, "map.txt") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if excludedROMFileExtensions[ext] {
			continue
		}

		displayName := stripROMDisplayTags(name)
		mapEntries = append(mapEntries, mapEntry{filename: name, displayName: displayName})
		displayCount[strings.ToLower(displayName)]++
	}

	if len(mapEntries) == 0 {
		return 0, 0, nil
	}

	for _, count := range displayCount {
		if count > 1 {
			collisions += count - 1
		}
	}

	sort.Slice(mapEntries, func(i, j int) bool {
		return mapEntries[i].filename < mapEntries[j].filename
	})

	if g.dryRun {
		emitVerbose(g, "curated", "convert", "dry-run map.txt", outputFields{
			"dir":        dstSystemDir,
			"entries":    len(mapEntries),
			"collisions": collisions,
		})
		return 1, collisions, nil
	}

	var b strings.Builder
	for _, e := range mapEntries {
		b.WriteString(e.filename)
		b.WriteByte('|')
		b.WriteString(e.displayName)
		b.WriteByte('\n')
	}

	mapPath := filepath.Join(dstSystemDir, "map.txt")
	if err := os.WriteFile(mapPath, []byte(b.String()), 0o644); err != nil {
		return 0, collisions, err
	}

	return 1, collisions, nil
}

// generateAllSystemMapTxt scans all system folders under romsDstRoot and writes
// a map.txt for each flat (non-tree-mode) system.
func generateAllSystemMapTxt(romsDstRoot string, g globalFlags) (files int, collisions int, err error) {
	systemDirs, err := os.ReadDir(romsDstRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	for _, entry := range systemDirs {
		if !entry.IsDir() {
			continue
		}
		folderName := entry.Name()
		systemKey := systemKeyFromFolderName(folderName)
		if systemModes[strings.ToLower(systemKey)] == systemModeTree {
			continue
		}

		systemDir := filepath.Join(romsDstRoot, folderName)
		f, c, mapErr := generateSystemMapTxt(systemDir, g)
		if mapErr != nil {
			return files, collisions, mapErr
		}
		files += f
		collisions += c
	}

	return files, collisions, nil
}

func ensureArcadeMap(romsSrc, romsDstRoot string, g globalFlags) (bool, error) {
	arcadeDst := filepath.Join(romsDstRoot, nextUISystemFolderName("ARCADE"), "map.txt")
	candidates := []string{
		filepath.Join(romsSrc, "ARCADE", "map.txt"),
		filepath.Join(romsSrc, "NEOGEO", "map.txt"),
		filepath.Join(romsSrc, "CPS3", "map.txt"),
	}

	for _, candidate := range candidates {
		if exists, err := fileExistsAt(candidate); err != nil {
			return false, err
		} else if !exists {
			continue
		}

		if g.dryRun {
			emitInfo(g, "curated", "convert", "dry-run arcade map", outputFields{"from": candidate, "to": arcadeDst})
			return true, nil
		}

		if err := fsutil.CopyFile(candidate, arcadeDst); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

type collectionEntry struct {
	Path        string
	SystemOrder int
	Name        string
}

func writeCollections(destination string, g globalFlags) (int, int, error) {
	collectionsRoot := filepath.Join(destination, "Collections")
	if g.dryRun {
		emitInfo(g, "curated", "convert", "dry-run collections root", outputFields{"path": collectionsRoot})
	} else {
		if err := fsutil.RemoveIfExists(collectionsRoot); err != nil {
			return 0, 0, err
		}
		if err := fsutil.EnsureDir(collectionsRoot); err != nil {
			return 0, 0, err
		}
	}

	entries, err := collectROMEntries(destination)
	if err != nil {
		return 0, 0, err
	}

	written := 0
	totalEntries := 0
	for _, spec := range collectionSpecs {
		matches := filterCollectionEntries(entries, spec)
		if len(matches) == 0 {
			continue
		}
		written++
		totalEntries += len(matches)

		if g.dryRun {
			emitInfo(g, "curated", "convert", "dry-run collection", outputFields{"name": spec.Name, "entries": len(matches)})
			continue
		}

		var b strings.Builder
		for i, m := range matches {
			if i > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(m.Path)
		}
		b.WriteByte('\n')

		out := filepath.Join(collectionsRoot, spec.Name+".txt")
		if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
			return written, totalEntries, err
		}
	}

	return written, totalEntries, nil
}

func collectROMEntries(destination string) ([]collectionEntry, error) {
	romsRoot := filepath.Join(destination, "Roms")
	entries := make([]collectionEntry, 0, 8192)

	err := filepath.WalkDir(romsRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}

		relToRoms, relErr := filepath.Rel(romsRoot, path)
		if relErr != nil {
			return relErr
		}
		if relToRoms == "." {
			return nil
		}

		parts := splitPath(relToRoms)
		if len(parts) == 0 {
			return nil
		}

		if d.IsDir() {
			name := strings.TrimSpace(d.Name())
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			if strings.EqualFold(name, "__MACOSX") {
				return filepath.SkipDir
			}
			if strings.EqualFold(name, ".media") || strings.EqualFold(name, ".hidden") {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		if excludedROMFilenames[strings.ToLower(strings.TrimSpace(name))] {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if excludedROMFileExtensions[ext] || strings.EqualFold(name, "map.txt") {
			return nil
		}

		relToDst, relDstErr := filepath.Rel(destination, path)
		if relDstErr != nil {
			return relDstErr
		}

		entries = append(entries, collectionEntry{
			Path:        "/" + filepath.ToSlash(relToDst),
			SystemOrder: parseSystemOrder(parts[0]),
			Name:        strings.ToLower(name),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func parseSystemOrder(systemFolder string) int {
	parts := strings.SplitN(strings.TrimSpace(systemFolder), ")", 2)
	if len(parts) < 2 {
		return 999
	}
	v, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 999
	}
	return v
}

func filterCollectionEntries(entries []collectionEntry, spec collectionSpec) []collectionEntry {
	matches := make([]collectionEntry, 0, 256)
	for _, e := range entries {
		if matchesCollectionName(e.Name, spec.Keywords) {
			matches = append(matches, e)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].SystemOrder != matches[j].SystemOrder {
			return matches[i].SystemOrder < matches[j].SystemOrder
		}
		if matches[i].Name != matches[j].Name {
			return matches[i].Name < matches[j].Name
		}
		return matches[i].Path < matches[j].Path
	})

	unique := make([]collectionEntry, 0, len(matches))
	seen := map[string]bool{}
	for _, m := range matches {
		if seen[m.Path] {
			continue
		}
		seen[m.Path] = true
		unique = append(unique, m)
	}

	return unique
}

func matchesCollectionName(name string, keywords []string) bool {
	lower := strings.ToLower(name)
	for _, keyword := range keywords {
		if containsBoundedKeyword(lower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func containsBoundedKeyword(haystack, keyword string) bool {
	if strings.TrimSpace(keyword) == "" {
		return false
	}

	start := 0
	for {
		idx := strings.Index(haystack[start:], keyword)
		if idx < 0 {
			return false
		}
		idx += start
		end := idx + len(keyword)

		leftOK := idx == 0 || !isWordChar(rune(haystack[idx-1]))
		rightOK := end >= len(haystack) || !isWordChar(rune(haystack[end]))
		if leftOK && rightOK {
			return true
		}

		start = idx + 1
	}
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func splitPath(rel string) []string {
	if strings.TrimSpace(rel) == "" {
		return nil
	}
	return strings.Split(rel, string(os.PathSeparator))
}

func requireDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("required directory does not exist: %s", path)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("required path is not a directory: %s", path)
	}
	return nil
}

func requireDirIfPresent(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("expected directory but found file: %s", path)
	}
	return nil
}

func fileExistsAt(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
