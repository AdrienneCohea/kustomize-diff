package git

import "testing"

func TestParseSymref(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
	}{
		{
			name: "main branch",
			out:  "ref: refs/heads/main\tHEAD\nabc123\tHEAD\n",
			want: "origin/main",
		},
		{
			name: "master branch",
			out:  "ref: refs/heads/master\tHEAD\ndef456\tHEAD\n",
			want: "origin/master",
		},
		{
			name: "branch with slashes",
			out:  "ref: refs/heads/release/v2\tHEAD\n789abc\tHEAD\n",
			want: "origin/release/v2",
		},
		{
			name: "no symref line",
			out:  "abc123\tHEAD\n",
			want: "origin/HEAD",
		},
		{
			name: "empty output",
			out:  "",
			want: "origin/HEAD",
		},
		{
			name: "symref not first line",
			out:  "abc123\tHEAD\nref: refs/heads/develop\tHEAD\n",
			want: "origin/develop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSymref([]byte(tt.out))
			if got != tt.want {
				t.Errorf("parseSymref(%q) = %q, want %q", tt.out, got, tt.want)
			}
		})
	}
}
