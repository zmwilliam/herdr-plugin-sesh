package picker

import (
	"strings"

	"github.com/fullerzz/herdr-plugin-sesh/internal/model"
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

func init() {
	// fzf's scoring tables must be initialized before matching. Keep the scheme
	// fixed so filtering and rendering can safely use the matcher concurrently.
	algo.Init("default")
}

var searchSeparators = strings.NewReplacer("-", " ", "_", " ", "/", " ", ".", " ")

func normalizeSearch(text string, separatorAware bool) string {
	if separatorAware {
		return searchSeparators.Replace(text)
	}
	return text
}

func fuzzyMatch(text, query string, separatorAware, positions bool) (algo.Result, []int) {
	chars := util.ToChars([]byte(normalizeSearch(text, separatorAware)))
	pattern := []rune(strings.ToLower(normalizeSearch(query, separatorAware)))
	result, indices := algo.FuzzyMatchV2(false, false, true, &chars, pattern, positions, nil)
	if indices == nil {
		return result, nil
	}
	return result, *indices
}

func searchName(s model.Session) string {
	if s.Worktree.Linked && s.Worktree.ParentWorkspaceName != "" {
		return s.Worktree.ParentWorkspaceName + " / " + s.Name
	}
	return s.Name
}

// matchName returns positions relative to the child's name even when the query
// spans its parent. Parent-only positions are outside the child's label.
func matchName(s model.Session, query string, separatorAware, positions bool) (algo.Result, []int) {
	result, indices := fuzzyMatch(s.Name, query, separatorAware, positions)
	if !s.Worktree.Linked || s.Worktree.ParentWorkspaceName == "" {
		return result, indices
	}
	contextual, contextIndices := fuzzyMatch(searchName(s), query, separatorAware, positions)
	if contextual.Start < 0 || result.Start >= 0 && contextual.Score <= result.Score {
		return result, indices
	}
	offset := len([]rune(s.Worktree.ParentWorkspaceName)) + len(" / ")
	for i := range contextIndices {
		contextIndices[i] -= offset
	}
	return contextual, contextIndices
}
