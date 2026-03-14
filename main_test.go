package main

import (
	"sort"
	"testing"
)

func TestUnion(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want []string // sorted for comparison
	}{
		{
			name: "both empty",
			a:    nil,
			b:    nil,
			want: []string{},
		},
		{
			name: "a only",
			a:    []string{"overlays/prod", "overlays/staging"},
			b:    nil,
			want: []string{"overlays/prod", "overlays/staging"},
		},
		{
			name: "b only",
			a:    nil,
			b:    []string{"overlays/prod"},
			want: []string{"overlays/prod"},
		},
		{
			name: "no overlap",
			a:    []string{"overlays/prod"},
			b:    []string{"overlays/staging"},
			want: []string{"overlays/prod", "overlays/staging"},
		},
		{
			name: "full overlap",
			a:    []string{"overlays/prod", "overlays/staging"},
			b:    []string{"overlays/prod", "overlays/staging"},
			want: []string{"overlays/prod", "overlays/staging"},
		},
		{
			name: "partial overlap",
			a:    []string{"base", "overlays/prod"},
			b:    []string{"overlays/prod", "overlays/staging"},
			want: []string{"base", "overlays/prod", "overlays/staging"},
		},
		{
			name: "deduplicates within a single slice via union",
			a:    []string{"overlays/prod", "overlays/prod"},
			b:    []string{"overlays/prod"},
			want: []string{"overlays/prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := union(tt.a, tt.b)
			sort.Strings(got)

			if len(got) != len(tt.want) {
				t.Errorf("union(%v, %v) = %v (len %d), want %v (len %d)",
					tt.a, tt.b, got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("union(%v, %v)[%d] = %q, want %q", tt.a, tt.b, i, got[i], tt.want[i])
				}
			}
		})
	}
}
