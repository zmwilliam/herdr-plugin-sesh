package picker

import (
	"testing"

	"github.com/fullerzz/herdr-plugin-sesh/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterAndCurrent(t *testing.T) {
	m := New([]model.Session{{Name: "api"}, {Name: "web"}})
	m.Filter("api")
	cur, ok := m.Current()
	require.True(t, ok)
	assert.Equal(t, "api", cur.Name)
}

func TestFilterMatchesOrderedFragments(t *testing.T) {
	m := New([]model.Session{{Name: "workspace launchpad"}, {Name: "unrelated"}})
	m.Filter("worklaun")
	require.Len(t, m.Filtered, 1)
	assert.Equal(t, "workspace launchpad", m.Filtered[0].Name)
	m.Filter("launwork")
	assert.Empty(t, m.Filtered)
}

func TestFilterMatchesParentAndWorktreeName(t *testing.T) {
	m := New([]model.Session{
		{Name: "launchpad", WorkspaceID: "launchpad"},
		{Name: "main", Path: "/trees/one", Worktree: model.WorktreeRelation{Linked: true, ParentWorkspaceName: "launchpad"}},
		{Name: "main", Path: "/trees/two", Worktree: model.WorktreeRelation{Linked: true, ParentWorkspaceName: "other"}},
	})
	m.Filter("lamain")
	require.Len(t, m.Filtered, 1)
	assert.Equal(t, "/trees/one", m.Filtered[0].Path)
}

func TestFilterMatchesFragmentsAcrossPathComponents(t *testing.T) {
	m := New([]model.Session{{Name: "launchpad", Path: "/work/launchpad"}})
	m.Filter("worklaun")
	require.Len(t, m.Filtered, 1)
	assert.Equal(t, "launchpad", m.Filtered[0].Name)
}

func TestFilterRanksFuzzyMatchesAndPreservesTies(t *testing.T) {
	m := New([]model.Session{
		{Name: "a-project-i"}, {Name: "api-worker"}, {Name: "api"}, {Name: "api-server"},
	})
	m.Filter("api")
	got := make([]string, 0, len(m.Filtered))
	for _, s := range m.Filtered {
		got = append(got, s.Name)
	}
	assert.Equal(t, []string{"api", "api-worker", "api-server", "a-project-i"}, got)
	m.Filter("")
	assert.Equal(t, m.All, m.Filtered)
}
func TestFilterDoesNotMutateSourceItems(t *testing.T) {
	m := New([]model.Session{{Name: "api"}, {Name: "web"}})
	m.Filter("web")
	m.Filter("")
	require.Len(t, m.Filtered, 2)
	require.Equal(t, "api", m.Filtered[0].Name)
	assert.Equal(t, "web", m.Filtered[1].Name)
}
func TestFilterResetsSelectionWhenQueryChanges(t *testing.T) {
	m := New([]model.Session{{Name: "api"}, {Name: "api-worker"}, {Name: "web"}})
	m.Move(1)
	m.Filter("api")
	cur, ok := m.Current()
	require.True(t, ok)
	assert.Equal(t, "api", cur.Name)
}
func TestFilterRanksNameMatchesBeforePathMatches(t *testing.T) {
	m := New([]model.Session{
		{Source: "zoxide", Path: "/work/api-unnamed"},
		{Source: "config", Name: "api-service", Path: "/work/service"},
		{Source: "herdr", Name: "api-both", Path: "/work/api-both"},
		{Source: "herdr", Name: "worker", Path: "/work/api-worker"},
		{Source: "config", Name: "api-web", Path: "/work/web"},
	})

	m.Filter("api")

	want := []string{"api-service", "api-both", "api-web", "", "worker"}
	require.Len(t, m.Filtered, len(want))
	for i, name := range want {
		require.Equal(t, name, m.Filtered[i].Name)
	}
}
func TestFilterKeepsNameMatchAheadOfPathOnlyWorktreeParent(t *testing.T) {
	m := New([]model.Session{
		{Source: "herdr", Name: "backend", Path: "/work/api", WorkspaceID: "w-parent"},
		{Source: "herdr", Name: "api-fix", Path: "/work/api-fix", WorkspaceID: "w-child", Worktree: model.WorktreeRelation{Linked: true, ParentWorkspaceID: "w-parent"}},
	})

	m.Filter("api")

	want := []string{"api-fix", "backend"}
	require.Len(t, m.Filtered, len(want))
	for i, name := range want {
		require.Equal(t, name, m.Filtered[i].Name)
	}
}
func TestFilterSelectsHomeDirectoryWhenQueryIsHome(t *testing.T) {
	t.Setenv("HOME", "/Users/zach")
	m := New([]model.Session{
		{Name: "home-manager", Path: "/tmp/home-manager"},
		{Name: "~", Path: "/Users/zach"},
	})
	m.Filter("home")
	cur, ok := m.Current()
	require.True(t, ok)
	assert.Equal(t, "~", cur.Name)
}

func TestZeroValueModelPrioritizesHome(t *testing.T) {
	t.Setenv("HOME", "/Users/zach")
	m := Model{All: []model.Session{
		{Name: "home-manager", Path: "/tmp/home-manager"},
		{Name: "~", Path: "/Users/zach"},
	}}

	m.Filter("home")

	cur, ok := m.Current()
	require.True(t, ok)
	assert.Equal(t, "~", cur.Name)
}

func TestFilterRanksActualHomePathBeforeMisleadingHomeName(t *testing.T) {
	t.Setenv("HOME", "/Users/zachfuller")
	m := New([]model.Session{
		{Name: "~", Path: "/tmp/home-archive"},
		{Name: "home-manager", Path: "/tmp/manager"},
		{Name: "~", Path: "/Users/zachfuller"},
	})

	m.Filter("home")

	want := []string{"/Users/zachfuller", "/tmp/manager", "/tmp/home-archive"}
	require.Len(t, m.Filtered, len(want))
	for i, path := range want {
		require.Equal(t, path, m.Filtered[i].Path)
	}
}

func TestFilterCanDisableHomePrioritization(t *testing.T) {
	t.Setenv("HOME", "/Users/zach")
	m := New([]model.Session{
		{Name: "path-before", Path: "/tmp/home-path"},
		{Name: "home-tools", Path: "/tmp/tools"},
		{Name: "home-manager", Path: "/tmp/manager"},
		{Name: "home-root", Path: "/Users/zach"},
		{Name: "path-after", Path: "/tmp/home-after"},
	})
	m.DisableHomePrioritization = true

	m.Filter("HOME")

	// Without the home override, path matches rank by fuzzy score; the alias
	// still includes the actual home directory even when its path doesn't match.
	want := []string{"home-tools", "home-manager", "path-before", "path-after", "home-root"}
	require.Len(t, m.Filtered, len(want))
	for i, name := range want {
		require.Equal(t, name, m.Filtered[i].Name)
	}
}
func TestSeparatorAwareMatch(t *testing.T) {
	assert.True(t, Match("my-api.service", "api service", true), "expected separator aware match")
}

func TestSeparatorAwareMatchNormalizesQuerySeparators(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hay   string
		query string
	}{
		{"query keeps the candidate's separator", "my-api.service", "api-service"},
		{"query separator differs from the candidate's", "my-api.service", "api/service"},
		{"query separator matches a path boundary", "/work/api-service", "work/api"},
		{"underscore query against a dashed candidate", "my-api-service", "api_service"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, Match(tc.hay, tc.query, true))
		})
	}
}

// Normalizing the query is strictly additive: the candidate never retains any
// of the four separators, so a query containing one could not match before.
func TestSeparatorAwareMatchIsAdditive(t *testing.T) {
	for _, tc := range []struct {
		hay   string
		query string
		want  bool
	}{
		{"my-api.service", "api", true},
		{"my-api.service", "nope", false},
		{"my-api.service", "api service", true},
		{"my-api.service", "service api", false},
		{"my-api.service", "", true},
	} {
		assert.Equalf(t, tc.want, Match(tc.hay, tc.query, true), "Match(%q, %q, true)", tc.hay, tc.query)
	}
}

func TestSeparatorUnawareMatchLeavesQueryLiteral(t *testing.T) {
	require.True(t, Match("my-api.service", "api.service", false), "expected literal match when separator awareness is off")
	assert.False(t, Match("my-api.service", "api service", false), "did not expect separator normalization when it is off")
}
