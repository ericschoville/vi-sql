package mcp

import (
	"testing"

	"github.com/kopecmaciej/vi-sql/internal/database"
)

// fakeResolver is a minimal ConnectionResolver for testing name/driver resolution.
type fakeResolver struct {
	conns []ConnInfo
}

func (f fakeResolver) ResolveDriver(name string) (database.Driver, bool) { return nil, false }
func (f fakeResolver) ListConnections() []ConnInfo                       { return f.conns }

func TestResolveConnName(t *testing.T) {
	conns := []ConnInfo{
		{Name: "dars", Active: false},
		{Name: "sis", Active: true},
	}

	tests := []struct {
		name  string
		input string
		conns ConnectionResolver
		want  string
	}{
		{"explicit name passes through", "dars", fakeResolver{conns}, "dars"},
		{"explicit name need not be active", "sis", fakeResolver{conns}, "sis"},
		{"empty resolves to active", "", fakeResolver{conns}, "sis"},
		{"empty with no active stays empty", "", fakeResolver{[]ConnInfo{{Name: "dars"}}}, ""},
		{"empty with no connections stays empty", "", fakeResolver{nil}, ""},
		{"nil resolver returns input", "dars", nil, "dars"},
		{"nil resolver empty stays empty", "", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{connections: tt.conns}
			if got := s.resolveConnName(tt.input); got != tt.want {
				t.Errorf("resolveConnName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
