package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandDisplaysHelp(t *testing.T) {
	output := &bytes.Buffer{}

	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	rootCmd.SetArgs([]string{})

	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	rendered := output.String()

	expectedParts := []string{
		"Available Commands:",
		"secret",
		"namespace",
		"shell",
		"exec",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(rendered, expected) {
			t.Fatalf(
				"root help does not contain %q\n\n%s",
				expected,
				rendered,
			)
		}
	}
}

func TestSecretCommandAliases(t *testing.T) {
	want := []string{"secrets", "sec"}

	if len(secretCmd.Aliases) != len(want) {
		t.Fatalf(
			"secret aliases = %v, want %v",
			secretCmd.Aliases,
			want,
		)
	}

	for index, alias := range want {
		if secretCmd.Aliases[index] != alias {
			t.Fatalf(
				"secret alias %d = %q, want %q",
				index,
				secretCmd.Aliases[index],
				alias,
			)
		}
	}
}
