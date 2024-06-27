package gitignore

import "testing"

func TestIgnore(t *testing.T) {
	patterns := []string{"bin/", "**.out", "venv/", "*.py[co]"}
	testCases := []struct {
		input string
		want  bool
		wantErr bool
	}{
		{"./bin/a.exe", true, false},
		{"./a.out", true, false},
		{"foo", false, false},
		{"./query/module.py", false, false},
		{"a.pyc", true, false},
		{"venv/module/ll.o", true, false},
	}

	matcher := NewGitignoreMatcher().FromPatterns(patterns)
	if err := matcher.Build(); err !=nil {
		panic(err)
	}
	for _, tc := range testCases {
		have, err := matcher.Matches(tc.input)
		if have != tc.want {
			t.Errorf("IsIgnored(\"%v\")=%v, wanted %v instead", tc.input, have, tc.want)
		}
		if err != nil && !tc.wantErr {
			t.Errorf("IsIgnored(\"%v\") gave err=%v despite not wanting it", tc.input, err)
		}
	}
}
