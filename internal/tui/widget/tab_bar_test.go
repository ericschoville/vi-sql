package widget

import (
	"testing"

	"github.com/kopecmaciej/tview"
	"github.com/stretchr/testify/assert"
)

type mockPrimitive struct{ tview.Box }

func (m *mockPrimitive) Render() {}

func TestTabBar_SwitchToTabByName(t *testing.T) {
	tests := []struct {
		name     string
		tabName  string
		tabKind  TabKind
		wantFind bool
	}{
		{"exact match", "users", KindStructure, true},
		{"same name wrong kind", "users", KindView, false},
		{"different name", "orders", KindStructure, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewTabBar()
			bar.AddTab("users", &mockPrimitive{}, KindStructure, true)
			bar.AddTab("posts", &mockPrimitive{}, KindTable, false)

			found := bar.SwitchToTabByName(tt.tabName, tt.tabKind)
			assert.Equal(t, tt.wantFind, found)
		})
	}
}

func TestTabBar_DynamicQueryTabs(t *testing.T) {
	bar := NewTabBar()
	q1 := &mockPrimitive{}
	q2 := &mockPrimitive{}
	// Query tabs with MCP IDs — the only ones that should be returned.
	bar.AddDynamicTabWithID("Query 1", "tab-1", KindQuery, q1)
	bar.AddDynamicTabWithID("Query 2", "tab-2", KindQuery, q2)
	// Query tab without an ID — not MCP-addressable, must be excluded.
	bar.AddDynamicTab("Query 3", &mockPrimitive{}, KindQuery)
	// Non-query tab with an ID — wrong kind, must be excluded.
	bar.AddDynamicTabWithID("users", "tab-struct", KindStructure, &mockPrimitive{})

	got := bar.DynamicQueryTabs()

	assert.Len(t, got, 2)
	assert.Same(t, q1, got["tab-1"])
	assert.Same(t, q2, got["tab-2"])
	assert.NotContains(t, got, "tab-struct")
}

func TestTabBar_DynamicQueryTabs_Empty(t *testing.T) {
	bar := NewTabBar()
	bar.AddTab("users", &mockPrimitive{}, KindStructure, true)
	assert.Empty(t, bar.DynamicQueryTabs())
}
