package kubernetes

import "testing"

func TestIsExecutableNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "container runtime executable missing",
			err:  testExecError("executable file not found in $PATH"),
			want: true,
		},
		{
			name: "path missing",
			err:  testExecError("stat /bin/bash: no such file or directory"),
			want: true,
		},
		{
			name: "permission denied",
			err:  testExecError("pods/exec is forbidden"),
			want: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := isExecutableNotFoundError(test.err); got != test.want {
				t.Fatalf("isExecutableNotFoundError() = %v, want %v", got, test.want)
			}
		})
	}
}

type testExecError string

func (err testExecError) Error() string { return string(err) }
