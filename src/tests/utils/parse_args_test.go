package tests

import (
	"reflect"
	"testing"

	"accretional.com/semantifly/utils"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string][]string
	}{
		{
			name: "Simple flag and non-flag arguments",
			args: []string{"--flag1", "value1", "arg1"},
			expected: map[string][]string{
				"flags":    {"--flag1", "value1"},
				"nonFlags": {"arg1"},
			},
		},
		{
			name: "Flags with equal sign",
			args: []string{"-flag1=value1", "arg1"},
			expected: map[string][]string{
				"flags":    {"-flag1", "value1"},
				"nonFlags": {"arg1"},
			},
		},
		{
			name: "Multiple flags and non-flag arguments",
			args: []string{"-flag1=value1", "-flag2", "value2", "arg1", "arg2"},
			expected: map[string][]string{
				"flags":    {"-flag1", "value1", "-flag2", "value2"},
				"nonFlags": {"arg1", "arg2"},
			},
		},
		{
			name: "No flags, only non-flag arguments",
			args: []string{"arg1", "arg2", "arg3"},
			expected: map[string][]string{
				"flags":    nil,
				"nonFlags": {"arg1", "arg2", "arg3"},
			},
		},
		{
			name: "Complicated different flags and order of arguments",
			args: []string{"--flag1", "value1", "arg1", "arg2", "-flag2=value2", "arg3", "-flag3", "value3"},
			expected: map[string][]string{
				"flags":    {"--flag1", "value1", "-flag2", "value2", "-flag3", "value3"},
				"nonFlags": {"arg1", "arg2", "arg3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ParseArgs(tt.args)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseArgs() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
