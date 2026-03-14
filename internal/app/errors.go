package app

import "strings"

// FormatCLIError returns a user-facing error message with practical hints.
func FormatCLIError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return ""
	}

	hints := errorHints(msg)
	if len(hints) == 0 {
		return msg
	}

	var b strings.Builder
	b.WriteString(msg)
	b.WriteString("\n")
	for _, hint := range hints {
		b.WriteString("  - ")
		b.WriteString(hint)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func errorHints(msg string) []string {
	lower := strings.ToLower(msg)
	hints := make([]string, 0, 3)

	switch {
	case strings.Contains(lower, "no command provided"):
		hints = append(hints,
			"Use 'retro-collection-tool help' to list commands.",
			"Run 'retro-collection-tool help <command>' for command-specific examples.",
		)
	case strings.Contains(lower, "unknown command:"):
		hints = append(hints,
			"Check available commands with 'retro-collection-tool help'.",
			"Command names are lowercase (for example: sync, hacks, bios, export).",
		)
	case strings.Contains(lower, "no config found"):
		hints = append(hints,
			"Pass an explicit config path with '--config <path>'.",
			"Or set RETRO_COLLECTION_TOOL_CONFIG to your config file.",
		)
	case strings.Contains(lower, "export requires --destination"):
		hints = append(hints,
			"Provide a destination path, for example '--destination /media/SDCARD'.",
			"See usage examples with 'retro-collection-tool help export'.",
		)
	case strings.Contains(lower, "cache requires subcommand"):
		hints = append(hints,
			"Use one of: 'retro-collection-tool cache clean' or 'retro-collection-tool cache path'.",
		)
	case strings.Contains(lower, "unknown cache subcommand"):
		hints = append(hints,
			"Valid cache subcommands are: clean, path.",
		)
	case strings.Contains(lower, "unexpected arguments"):
		hints = append(hints,
			"Run 'retro-collection-tool help <command>' to verify supported flags.",
		)
	case strings.Contains(lower, "bios strict mode failed"):
		hints = append(hints,
			"Re-run with '--verbose' to list required missing BIOS files.",
			"If you only want best-effort imports, omit '--strict'.",
		)
	case strings.Contains(lower, "workflow disabled in config.features"):
		hints = append(hints,
			"Enable the matching feature flag in your config file under 'features'.",
		)
	case strings.Contains(lower, "compression requested") && strings.Contains(lower, "allow_compression_zip"):
		hints = append(hints,
			"Set 'igir.allow_compression_zip: true' in config or remove '--compress'.",
		)
	}

	return hints
}
