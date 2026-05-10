package main

import "testing"

func TestCliFlags(t *testing.T) {

	tests := []struct {
		Name        string
		Args        []string
		Expected    cliFlags
		ExpectError bool
	}{
		{"-h", []string{"-h"}, cliFlags{}, true},
		{"--help", []string{"--help"}, cliFlags{}, true},

		{"-r", []string{"-r"}, cliFlags{dbReset: true}, false},
		{"--reset", []string{"--reset"}, cliFlags{dbReset: true}, false},

		{"-ru", []string{"-ru"}, cliFlags{usersReset: true}, false},
		{"-reset-users", []string{"--reset-users"}, cliFlags{usersReset: true}, false},

		{"-rs", []string{"-rs"}, cliFlags{clearSessions: true}, false},
		{"--reset-sessions", []string{"--reset-sessions"}, cliFlags{clearSessions: true}, false},

		{"-cl", []string{"-cl"}, cliFlags{clearLibrary: true}, false},
		{"--clear-library", []string{"-cl"}, cliFlags{clearLibrary: true}, false},

		{"--cd", []string{"-cd"}, cliFlags{clearDownloads: true}, false},
		{"--clear-downloads", []string{"--clear-downloads"}, cliFlags{clearDownloads: true}, false},

		{"-t", []string{"-t"}, cliFlags{dbTestData: true, dbReset: true}, false},
		{"--test-dataset", []string{"--test-dataset"}, cliFlags{dbTestData: true, dbReset: true}, false},
		{"-t l", []string{"-t", "l"}, cliFlags{dbTestData: true, dbReset: true, libTestData: true}, false},
		{"-t test-library", []string{"-t", "test-library"}, cliFlags{dbTestData: true, dbReset: true, libTestData: true}, false},
		{"-t d", []string{"-t", "d"}, cliFlags{dbTestData: true, dbReset: true, downTestData: true}, false},
		{"-t test-downloads", []string{"-t", "test-downloads"}, cliFlags{dbTestData: true, dbReset: true, downTestData: true}, false},

		{"Clear Directories (-cl -cd)", []string{"-cl", "-cd"}, cliFlags{clearLibrary: true, clearDownloads: true}, false},
		{"Unknown", []string{"cd"}, cliFlags{}, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			flags, err := getFlags(test.Args)
			if err != nil && !test.ExpectError {
				t.Error(err)
				return
			}

			if flags != test.Expected {
				t.Errorf("Returned flags do not match expected flags -> Expected: %v Got: %v", test.Expected, flags)
			}
		})
	}

}
