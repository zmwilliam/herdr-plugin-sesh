package sources

import (
	"testing"

	"github.com/fullerzz/herdr-plugin-sesh/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNameZoxidePreservesExplicitNames(t *testing.T) {
	sessions := []model.Session{
		{Source: "config", Name: "main", Path: "/configured"},
		{Source: "herdr", Name: "project/main", Path: "/workspace"},
		{Source: "zoxide", Name: "main", Path: "/projects/project/main"},
		{Source: "zoxide", Name: "main", Path: "/other/main"},
		{Source: "zoxide", Name: "/", Path: "/"},
	}
	NameZoxide(sessions, 1)
	assert.Equal(t, "main", sessions[0].Name)
	assert.Equal(t, "project/main", sessions[1].Name)
	assert.Equal(t, "projects/project/main", sessions[2].Name)
	assert.Equal(t, "other/main", sessions[3].Name)
	assert.Equal(t, "/", sessions[4].Name)
}

func TestNameZoxideDistinguishesPathSuffixes(t *testing.T) {
	sessions := []model.Session{
		{Source: "zoxide", Path: "/main"},
		{Source: "zoxide", Path: "/project/main"},
	}
	NameZoxide(sessions, 1)
	assert.Equal(t, "/main", sessions[0].Name)
	assert.Equal(t, "project/main", sessions[1].Name)
}
