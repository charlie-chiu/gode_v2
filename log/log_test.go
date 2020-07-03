package log

import "testing"

type TestCase struct {
	logLevelString string
	wantLevel      int
}

type TestCases []TestCase

func TestParseLogLevel(t *testing.T) {
	// everything invalid parsing to nothing
	testCases := TestCases{
		{"debug", Debug},
		{"Debug", Debug},
		{"DEBUG", Debug},
		{"info", Info},
		{"NOTICE", Notice},
		{"hello", Nothing},
		{"world", Nothing},
		{"123454", Nothing},
		{"nada", Nothing},
		{"1", Nothing},
	}

	for _, tc := range testCases {
		gotLevel := ParseLogLevel(tc.logLevelString)
		if gotLevel != tc.wantLevel {
			t.Errorf("parsing log level %q: want %s , got %s", tc.logLevelString, logText[tc.wantLevel], logText[gotLevel])
		}
	}
}
