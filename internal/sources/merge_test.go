package sources

import (
	"context"
	"testing"

	"github.com/fullerzz/herdr-plugin-sesh/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticSource struct {
	name     string
	sessions []model.Session
}

func (s staticSource) Name() string { return s.name }
func (s staticSource) List(context.Context) (model.Sessions, error) {
	out := model.NewSessions()
	for _, x := range s.sessions {
		out.Add(x)
	}
	return out, nil
}
func TestMergeOrderBlacklistDedupe(t *testing.T) {
	srcs := []Source{staticSource{"zoxide", []model.Session{{Source: "zoxide", Name: "api", Path: "/z"}}}, staticSource{"config", []model.Session{{Source: "config", Name: "api", Path: "/c"}, {Source: "config", Name: "scratch", Path: "/s"}}}}
	got, err := Merge(context.Background(), srcs, []string{"config", "zoxide"}, []string{"scratch"}, false, true)
	require.NoError(t, err)
	os := got.Ordered()
	require.Len(t, os, 2)
	require.Equal(t, "config", os[0].Source)
	assert.Equal(t, "api", os[0].Name)
	assert.Equal(t, "/z", os[1].Path)
}
func TestParseZoxideLine(t *testing.T) {
	s, ok := ParseZoxideLine("42.5 /tmp/my app")
	require.True(t, ok)
	require.InDelta(t, 42.5, s.Score, 0)
	require.Equal(t, "/tmp/my app", s.Path)
	assert.Equal(t, "my app", s.Name)
}

func TestMergePreservesDistinctDirectoryPaths(t *testing.T) {
	srcs := []Source{
		staticSource{"config", []model.Session{{Source: "config", Name: "main", Path: "/projects/project-a/main"}}},
		staticSource{"zoxide", []model.Session{
			{Source: "zoxide", Name: "main", Path: "/projects/project-a/main"},
			{Source: "zoxide", Name: "main", Path: "/projects/project-b/main"},
		}},
	}
	got, err := Merge(context.Background(), srcs, []string{"config", "zoxide"}, nil, false, true)
	require.NoError(t, err)
	ordered := got.Ordered()
	require.Len(t, ordered, 2)
	assert.Equal(t, "config", ordered[0].Source)
	assert.Equal(t, "/projects/project-b/main", ordered[1].Path)
}
