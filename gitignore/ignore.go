package gitignore

import (
	"io"
	"os"
	"strings"

	"github.com/moby/patternmatcher"
)

// Represents a GitignoreMatcher.
type GitignoreMatcher struct {
	patterns []string
	matcher  *patternmatcher.PatternMatcher
}

// Returns a pointer to the [GitignoreMatcher] with no patterns.
func NewGitignoreMatcher() *GitignoreMatcher {
	return &GitignoreMatcher{patterns: []string{}}
}

// Adds the given patterns to the existing [GitignoreMatcher].
func (gm *GitignoreMatcher) FromPatterns(patterns []string) *GitignoreMatcher {
	gm.patterns = append(gm.patterns, patterns...)
	return gm
}

// Adds the given patterns to the existing [GitignoreMatcher] by reading from the given reader.
func (gm *GitignoreMatcher) FromReader(r io.Reader) (*GitignoreMatcher, error) {
	var buf []byte
	_, err := r.Read(buf)
	if err != nil {
		return nil, err
	}

	stringBuf := string(buf)
	patterns := strings.Split(stringBuf, "\n")
	gm.patterns = append(gm.patterns, patterns...)
	return gm, nil
}

// Adds the given patterns to the existing [GitignoreMatcher] by reading from the given file.
func (gm *GitignoreMatcher) FromFile(filename string) (*GitignoreMatcher, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return gm.FromReader(file)
}

// Build must be called before using the GitignoreMatcher. 
// It initializes the [patternmatcher.PatternMatcher] struct with the given patterns.
func (gm *GitignoreMatcher) Build() error {
	m, e := patternmatcher.New(gm.patterns)
	if e != nil {
		return e
	}
	gm.matcher = m
	return nil
}

// Returns a boolean value indicating whether the given filepath would be ignored, as per the gitignore spec.
func (gm *GitignoreMatcher) Matches(someFilePath string) (bool, error) {
	if gm.matcher == nil {
		panic("Build method not called on GitignoreMatcher")
	}
	return gm.matcher.MatchesOrParentMatches(someFilePath)
}
