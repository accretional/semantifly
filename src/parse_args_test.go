package main

import (
	"errors"
	"flag"
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		expectedFlags        []string
		expectedNonFlags     []string
		expectError          bool
		expectedErrorMessage error
	}{
		{
			name:             "Simple flag and non-flag arguments",
			args:             []string{"--flag1", "value1", "arg1"},
			expectedFlags:    []string{"--flag1", "value1"},
			expectedNonFlags: []string{"arg1"},
			expectError:      false,
		},
		{
			name:             "Flags with equal sign",
			args:             []string{"-flag1=value1", "arg1"},
			expectedFlags:    []string{"-flag1=value1"},
			expectedNonFlags: []string{"arg1"},
			expectError:      false,
		},
		{
			name:             "Boolean flag",
			args:             []string{"-boolflag1", "arg1"},
			expectedFlags:    []string{"-boolflag1"},
			expectedNonFlags: []string{"arg1"},
			expectError:      false,
		},
		{
			name:             "Boolean flag v2",
			args:             []string{"arg1", "--boolflag1"},
			expectedFlags:    []string{"--boolflag1"},
			expectedNonFlags: []string{"arg1"},
			expectError:      false,
		},
		{
			name:             "Complicated bool flags",
			args:             []string{"-boolflag1=true", "arg1", "--boolflag2"},
			expectedFlags:    []string{"-boolflag1=true", "--boolflag2"},
			expectedNonFlags: []string{"arg1"},
			expectError:      false,
		},
		{
			name:             "Multiple flags and non-flag arguments",
			args:             []string{"--flag1=value1", "arg1", "arg2", "-flag2", "value2"},
			expectedFlags:    []string{"--flag1=value1", "-flag2", "value2"},
			expectedNonFlags: []string{"arg1", "arg2"},
			expectError:      false,
		},
		{
			name:             "No flags, only non-flag arguments",
			args:             []string{"arg1", "arg2", "arg3"},
			expectedFlags:    nil,
			expectedNonFlags: []string{"arg1", "arg2", "arg3"},
			expectError:      false,
		},
		{
			name:             "Only flags, no non-flag arguments",
			args:             []string{"--flag1=value1", "--flag2", "value2"},
			expectedFlags:    []string{"--flag1=value1", "--flag2", "value2"},
			expectedNonFlags: nil,
			expectError:      false,
		},
		{
			name:             "Complicated test: different types of flags and order of arguments",
			args:             []string{"--flag1", "value1", "arg1", "arg2", "-boolflag1", "-flag2=value2", "arg3", "-flag3", "value3", "--boolflag2=true"},
			expectedFlags:    []string{"--flag1", "value1", "-boolflag1", "-flag2=value2", "-flag3", "value3", "--boolflag2=true"},
			expectedNonFlags: []string{"arg1", "arg2", "arg3"},
			expectError:      false,
		},
		{
			name:             "No args",
			args:             []string{},
			expectedFlags:    nil,
			expectedNonFlags: nil,
			expectError:      false,
		},
	}

	// should work for all tests
	cmd := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.String("flag1", "", "test flag 1")
	cmd.String("flag2", "", "test flag 2")
	cmd.String("flag3", "", "test flag 3")
	cmd.Bool("boolflag1", false, "test bool flag 1")
	cmd.Bool("boolflag2", true, "test bool flag 2")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			flags, nonFlags, err := parseArgs(tt.args, cmd)

			if tt.expectError && err == nil {
				t.Errorf("parseArgs() expected an error, but got none")
			} else if tt.expectError && errors.Is(err, tt.expectedErrorMessage) {
				t.Errorf("parseArgs() returned error message = %v, expected %v", err, tt.expectedErrorMessage)
			} else if !tt.expectError && err != nil {
				t.Errorf("parseArgs() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(flags, tt.expectedFlags) {
				t.Errorf("parseArgs() flags = %v, expected %v", flags, tt.expectedFlags)
			}

			if !reflect.DeepEqual(nonFlags, tt.expectedNonFlags) {
				t.Errorf("parseArgs() nonFlags = %v, expected %v", nonFlags, tt.expectedNonFlags)
			}
		})
	}
}
