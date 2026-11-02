package picker

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fullerzz/herdr-plugin-sesh/internal/model"
)

type Model struct {
	All                       []model.Session
	Filtered                  []model.Session
	Query                     string
	Selected                  int
	SeparatorAware            bool
	DisableHomePrioritization bool
}

func New(items []model.Session) Model {
	return Model{
		All:      append([]model.Session(nil), items...),
		Filtered: append([]model.Session(nil), items...),
	}
}
func (m *Model) Filter(q string) {
	queryChanged := q != m.Query
	m.Query = q
	m.Filtered = m.Filtered[:0]
	type rankedSession struct {
		session model.Session
		group   int
		score   int
		exact   bool
	}
	matches := make([]rankedSession, 0, len(m.All))
	homeQuery := strings.EqualFold(q, "home")
	for _, s := range m.All {
		if q == "" {
			m.Filtered = append(m.Filtered, s)
			continue
		}
		name, _ := matchName(s, q, m.SeparatorAware, false)
		if homeQuery && isHomePath(s.Path) {
			if !m.DisableHomePrioritization {
				matches = append(matches, rankedSession{session: s})
			} else {
				path, _ := fuzzyMatch(s.Path, q, m.SeparatorAware, false)
				matches = append(matches, rankedSession{session: s, group: 2, score: path.Score})
			}
			continue
		}
		if name.Start >= 0 {
			matches = append(matches, rankedSession{session: s, group: 1, score: name.Score, exact: strings.EqualFold(normalizeSearch(s.Name, m.SeparatorAware), normalizeSearch(q, m.SeparatorAware))})
		} else {
			path, _ := fuzzyMatch(s.Path, q, m.SeparatorAware, false)
			if path.Start >= 0 {
				matches = append(matches, rankedSession{session: s, group: 2, score: path.Score})
			}
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		if a.group != b.group {
			return a.group < b.group
		}
		if a.exact != b.exact {
			return a.exact
		}
		return a.score > b.score
	})
	for _, match := range matches {
		m.Filtered = append(m.Filtered, match.session)
	}
	if queryChanged {
		m.Selected = 0
	}
	m.clampSelected()
}
func (m *Model) Move(delta int) {
	m.Selected += delta
	m.clampSelected()
}
func (m *Model) UpdateAgentStatuses(statuses map[string]string) {
	for i := range m.All {
		if m.All[i].WorkspaceID != "" {
			m.All[i].AgentStatus = statuses[m.All[i].WorkspaceID]
		}
	}
	m.Filter(m.Query)
}
func (m *Model) clampSelected() {
	if m.Selected >= len(m.Filtered) {
		m.Selected = len(m.Filtered) - 1
	}
	if m.Selected < 0 {
		m.Selected = 0
	}
}
func (m *Model) Current() (model.Session, bool) {
	if len(m.Filtered) == 0 || m.Selected < 0 || m.Selected >= len(m.Filtered) {
		return model.Session{}, false
	}
	return m.Filtered[m.Selected], true
}
func Match(s, q string, sep bool) bool {
	result, _ := fuzzyMatch(s, q, sep, false)
	return result.Start >= 0
}
func isHomePath(p string) bool {
	if p == "" {
		return false
	}
	cleaned := filepath.Clean(p)
	if cleaned == "~" {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return false
	}
	return strings.EqualFold(cleaned, filepath.Clean(home))
}
func isHomeSession(s model.Session) bool {
	return isHomePath(s.Name) || isHomePath(s.Path)
}
