package platform

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
)

func ExpandSystems(requested []string, all bool, cfg *config.Config) ([]string, error) {
	if all {
		return cfg.EnabledSystems(), nil
	}
	if len(requested) == 0 {
		return nil, fmt.Errorf("no systems selected; use --systems or --all-systems")
	}

	normalized := make([]string, 0, len(requested))
	seen := map[string]bool{}
	for _, raw := range requested {
		for _, p := range strings.Split(raw, ",") {
			s := strings.TrimSpace(strings.ToLower(p))
			if s == "" {
				continue
			}

			resolved, ok := resolveSystemToken(s, cfg)
			if !ok {
				return nil, fmt.Errorf("unsupported system: %s", s)
			}

			if seen[resolved] {
				continue
			}
			if !cfg.Systems[resolved].Enabled {
				return nil, fmt.Errorf("system disabled in config: %s", resolved)
			}
			seen[resolved] = true
			normalized = append(normalized, resolved)
		}
	}
	sort.Strings(normalized)
	return normalized, nil
}

var explicitSystemAliases = map[string]string{
	"ngp":  "neo-geo-pocket",
	"ngpc": "neo-geo-pocket-color",
}

func resolveSystemToken(token string, cfg *config.Config) (string, bool) {
	if _, ok := cfg.Systems[token]; ok {
		return token, true
	}

	if alias, ok := explicitSystemAliases[token]; ok {
		if _, exists := cfg.Systems[alias]; exists {
			return alias, true
		}
	}

	normalized := normalizeSystemToken(token)
	if normalized == "" {
		return "", false
	}

	for key := range cfg.Systems {
		if normalizeSystemToken(key) == normalized {
			return key, true
		}
	}

	return "", false
}

func normalizeSystemToken(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
