package sources

import (
	"path/filepath"
	"strings"

	"github.com/fullerzz/herdr-plugin-sesh/internal/model"
)

// NameZoxide uses at least components trailing path components, expanding
// ambiguous labels without changing explicitly named config or Herdr sessions.
func NameZoxide(sessions []model.Session, components int) {
	components = max(1, components)
	parts := make([][]string, len(sessions))
	depths := make([]int, len(sessions))
	for i, session := range sessions {
		if session.Source != "zoxide" || session.Path == "" {
			continue
		}
		parts[i] = strings.Split(filepath.Clean(session.Path), string(filepath.Separator))
		if filepath.Clean(session.Path) == string(filepath.Separator) {
			parts[i] = []string{string(filepath.Separator)}
		}
		available := len(parts[i])
		if parts[i][0] == "" {
			available-- // The leading separator is not a path component.
		}
		depths[i] = min(components, available)
		sessions[i].Name = strings.Join(parts[i][len(parts[i])-depths[i]:], string(filepath.Separator))
	}
	for {
		paths := map[string]string{}
		ambiguous := map[string]bool{}
		for _, session := range sessions {
			path := session.Path
			if path != "" {
				path = filepath.Clean(path)
			}
			if previous, ok := paths[session.Name]; ok && previous != path {
				ambiguous[session.Name] = true
			}
			paths[session.Name] = path
		}
		changed := false
		for i := range sessions {
			if !ambiguous[sessions[i].Name] || depths[i] == len(parts[i]) {
				continue
			}
			depths[i]++
			sessions[i].Name = strings.Join(parts[i][len(parts[i])-depths[i]:], string(filepath.Separator))
			changed = true
		}
		if !changed {
			return
		}
	}
}
