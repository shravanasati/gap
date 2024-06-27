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
	// todo add a base dir component
	return &GitignoreMatcher{patterns: []string{}}
}

// Adds the given patterns to the existing [GitignoreMatcher].
func (gm *GitignoreMatcher) FromPatterns(patterns []string) *GitignoreMatcher {
	gm.patterns = append(gm.patterns, patterns...)
	return gm
}

// Adds the given patterns to the existing [GitignoreMatcher] by reading from the given reader.
func (gm *GitignoreMatcher) FromReader(r io.Reader) (*GitignoreMatcher, error) {
	buf := make([]byte, 1024)
	var builder strings.Builder
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		if _, err := builder.Write(buf); err != nil {
			panic("unable to write to strings.Builder in FromReader: " + err.Error())
		}
	}

	stringBuf := builder.String()
	patterns := strings.Split(stringBuf, "\n")
	gm.patterns = append(gm.patterns, patterns...)
	return gm, nil
}

// Adds the given patterns to the existing [GitignoreMatcher] by reading from the given file.
func (gm *GitignoreMatcher) FromFile(filename string) (*GitignoreMatcher, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	dataStr := string(data)
	patterns := strings.Split(dataStr, "\n")
	gm.patterns = append(gm.patterns, patterns...)
	return gm, nil
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
		panic("Build method either not called on GitignoreMatcher, or some constructor failed")
	}
	return gm.matcher.MatchesOrParentMatches(someFilePath)
}
