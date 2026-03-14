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
)

type outputFields map[string]any

func (g globalFlags) isJSONOutput() bool {
	return strings.EqualFold(strings.TrimSpace(g.outputMode), outputModeJSON)
}

func emitInfo(g globalFlags, command, phase, message string, fields outputFields) {
	emitLine(g, "info", command, phase, message, fields)
}

func emitError(g globalFlags, command, phase, message string, fields outputFields) {
	emitLine(g, "error", command, phase, message, fields)
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

	line := strings.TrimSpace(message)
	if len(fields) > 0 {
		line = line + " " + formatFields(fields)
	}
	if strings.TrimSpace(phase) == "" {
		fmt.Printf("[%s] %s\n", strings.TrimSpace(command), line)
		return
	}
	fmt.Printf("[%s] [%s] %s\n", strings.TrimSpace(command), strings.TrimSpace(phase), line)
}

func formatFields(fields outputFields) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, fields[k]))
	}
	return strings.Join(parts, " ")
}
