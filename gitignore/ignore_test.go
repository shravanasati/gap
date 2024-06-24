package gitignore

import "testing"

func TestIgnore(t *testing.T) {
	patterns := []string{"bin/", "*.out", "venv/", "*.py[co]"}
	testCases := []struct{
		input string
		want bool
	}{
		{"./bin/a.exe", true},
		{"./something/a.out", true},
		{"foo", false},
		{"./query/module.py", false},
		{"a.pyc", true},
		{"venv/module/ll.o", true},
	}

	matcher := NewGitignoreMatcher().FromPatterns(patterns)
	for _, tc := range testCases {
		have := matcher.IsIgnored(tc.input)
		if have != tc.want {
			t.Errorf("IsIgnored(\"%v\")=%v, wanted %v instead", tc.input, have, tc.want)
		}
	}
}
