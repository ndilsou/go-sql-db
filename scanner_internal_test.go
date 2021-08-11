package sql

import (
	"testing"
	"unicode"
)

func Test_isWhitespace(t *testing.T) {
	tests := []struct {
		name string
		ch   rune
		want bool
	}{
		{
			name: "space",
			ch:   ' ',
			want: true,
		},
		{
			name: "tab",
			ch:   '\t',
			want: true,
		},
		{
			name: "newline",
			ch:   '\n',
			want: true,
		},
		{
			name: "alpha",
			ch:   'B',
			want: false,
		},
		{
			name: "numeric",
			ch:   '1',
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWhitespace(tt.ch); got != tt.want {
				t.Errorf("isWhitespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isLetter(t *testing.T) {
	tests := []struct {
		name string
		ch   rune
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "newline",
			ch:   '\n',
			want: false,
		},
		{
			name: "caps alpha A",
			ch:   'A',
			want: true,
		},
		{
			name: "caps alpha Z",
			ch:   'Z',
			want: true,
		},
		{
			name: "caps alpha",
			ch:   'B',
			want: true,
		},
		{
			name: "alpha a",
			ch:   'a',
			want: true,
		},
		{
			name: "alpha z",
			ch:   'z',
			want: true,
		},
		{
			name: "alpha",
			ch:   'v',
			want: true,
		},
		{
			name: "numeric",
			ch:   '1',
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unicode.IsLetter(tt.ch); got != tt.want {
				t.Errorf("isLetter() = %v, want %v", got, tt.want)
			}
		})
	}
}
