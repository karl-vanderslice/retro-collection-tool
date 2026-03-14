package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	outputModeHuman = "human"
	outputModeJSON  = "json"

	colorModeAuto   = "auto"
	colorModeAlways = "always"
	colorModeNever  = "never"
)

type outputFields map[string]any

func (g globalFlags) isJSONOutput() bool {
	return strings.EqualFold(strings.TrimSpace(g.outputMode), outputModeJSON)
}

func (g globalFlags) usesColor() bool {
	if g.isJSONOutput() {
		return false
	}
	mode := strings.ToLower(strings.TrimSpace(g.colorMode))
	if mode == "" {
		mode = colorModeAuto
	}
	if shouldDisableColorFromEnv() {
		return false
	}
	switch mode {
	case colorModeAlways:
		return true
	case colorModeNever:
		return false
	default:
		return stdoutIsTerminal()
	}
}

func emitInfo(g globalFlags, command, phase, message string, fields outputFields) {
	emitLine(g, "info", command, phase, message, fields)
}

func emitWarn(g globalFlags, command, phase, message string, fields outputFields) {
	emitLine(g, "warn", command, phase, message, fields)
}

func emitError(g globalFlags, command, phase, message string, fields outputFields) {
	emitLine(g, "error", command, phase, message, fields)
}

func emitVerbose(g globalFlags, command, phase, message string, fields outputFields) {
	if !g.verbose {
		return
	}
	emitInfo(g, command, phase, message, fields)
}

func emitLine(g globalFlags, level, command, phase, message string, fields outputFields) {
	if g.isJSONOutput() {
		event := map[string]any{
			"time":    time.Now().UTC().Format(time.RFC3339),
			"level":   strings.TrimSpace(level),
			"command": strings.TrimSpace(command),
			"message": strings.TrimSpace(message),
		}
		if strings.TrimSpace(phase) != "" {
			event["phase"] = strings.TrimSpace(phase)
		}
		for k, v := range fields {
			event[k] = v
		}
		b, err := json.Marshal(event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "json output error: %v\n", err)
			return
		}
		fmt.Println(string(b))
		return
	}

	useColor := g.usesColor()
	prefix := formatHumanPrefix(command, phase, level, useColor)
	line := strings.TrimSpace(message)
	if len(fields) > 0 {
		line = line + "  " + formatHumanFields(fields, useColor)
	}
	if prefix != "" {
		fmt.Printf("%s %s\n", prefix, line)
	} else {
		fmt.Println(line)
	}
}

func formatHumanPrefix(command, phase, level string, color bool) string {
	parts := make([]string, 0, 3)
	if cmd := strings.ToUpper(strings.TrimSpace(command)); cmd != "" {
		parts = append(parts, styleBadge(cmd, "34", color))
	}
	if ph := strings.ToUpper(strings.TrimSpace(phase)); ph != "" {
		parts = append(parts, styleBadge(ph, "36", color))
	}
	if lvl := strings.ToLower(strings.TrimSpace(level)); lvl != "" {
		parts = append(parts, styleBadge(strings.ToUpper(lvl), levelColorCode(lvl), color))
	}
	return strings.Join(parts, " ")
}

func levelColorCode(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "error":
		return "31"
	case "warn":
		return "33"
	default:
		return "32"
	}
}

func styleBadge(text, colorCode string, color bool) string {
	t := "[" + text + "]"
	if !color {
		return t
	}
	return "\033[1;" + colorCode + "m" + t + "\033[0m"
}

func styleDim(text string, color bool) string {
	if !color {
		return text
	}
	return "\033[2m" + text + "\033[0m"
}

func formatHumanFields(fields outputFields, color bool) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", styleDim(k, color), fields[k]))
	}
	return strings.Join(parts, " ")
}

func shouldDisableColorFromEnv() bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("CLICOLOR")) == "0" {
		return true
	}
	return false
}
