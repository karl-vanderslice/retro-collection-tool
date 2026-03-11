package platform

import (
	"fmt"
	"sort"
	"strings"

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
			if s == "" || seen[s] {
				continue
			}
			if _, ok := cfg.Systems[s]; !ok {
				return nil, fmt.Errorf("unsupported system: %s", s)
			}
			if !cfg.Systems[s].Enabled {
				return nil, fmt.Errorf("system disabled in config: %s", s)
			}
			seen[s] = true
			normalized = append(normalized, s)
		}
	}
	sort.Strings(normalized)
	return normalized, nil
}
