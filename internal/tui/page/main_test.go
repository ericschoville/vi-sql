package page

import (
	"context"
	"testing"

	"github.com/kopecmaciej/vi-sql/internal/database"
	"github.com/kopecmaciej/vi-sql/internal/manager"
	"github.com/kopecmaciej/vi-sql/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newMainTestDriver returns a MockDriver covering the calls Main makes while
// switching connections: SchemaTree.Render → ListSchemas, its CreateTableModal
// init → CommonDataTypes, and the TopBar health monitor's background Ping.
func newMainTestDriver() *testutil.MockDriver {
	m := &testutil.MockDriver{}
	m.On("ListSchemas", context.Background(), "").Return([]database.Schema{}, nil)
	m.On("CommonDataTypes").Return([]string{"TEXT", "INTEGER"})
	m.On("Ping", mock.Anything).Return(nil)
	return m
}

func newMainForTest(t *testing.T) (*Main, *manager.TabRegistry) {
	t.Helper()
	app, _ := testutil.NewTestApp(t)
	app.SetDriver(newMainTestDriver())
	app.SetFormatter(database.DefaultFormatter{})

	m := NewMain()
	require.NoError(t, m.Init(app))
	reg := manager.NewTabRegistry()
	m.SetRegistry(reg)
	return m, reg
}

// Only the active workspace's MCP-addressable query tabs stay in the shared
// TabRegistry; switching connections unregisters the outgoing ones and
// re-registers the incoming ones.
func TestMain_SwitchToConnection_RegistryFollowsActiveWorkspace(t *testing.T) {
	m, reg := newMainForTest(t)
	driver := newMainTestDriver()

	// Connection A active with one MCP-addressable query tab.
	m.activeConn = "A"
	m.openNewQueryTabWithRequest(manager.OpenQueryTabRequest{TabID: "a-tab", Name: "qa", Query: "select 1"})
	require.True(t, reg.Exists("a-tab"))

	// Switch to a fresh connection B — A's tab must be unregistered.
	m.SwitchToConnection("B", driver)
	assert.False(t, reg.Exists("a-tab"), "A's tab must be unregistered after switching away")

	// Add an MCP tab on B.
	m.openNewQueryTabWithRequest(manager.OpenQueryTabRequest{TabID: "b-tab", Name: "qb", Query: "select 2"})
	require.True(t, reg.Exists("b-tab"))

	// Switch back to A — its tab returns, B's goes away.
	m.SwitchToConnection("A", driver)
	assert.True(t, reg.Exists("a-tab"), "A's tab must be re-registered on switch back")
	assert.False(t, reg.Exists("b-tab"), "B's tab must be unregistered after switching away")

	// The re-registered setter/getter target A's live tab.
	require.True(t, reg.SetText("a-tab", "updated"))
	got, ok := reg.GetText("a-tab")
	require.True(t, ok)
	assert.Equal(t, "updated", got)
}

// Closing the active connection with none remaining has no switch to piggyback
// on, so DropWorkspace must unregister its tabs directly.
func TestMain_DropWorkspace_UnregistersActiveTabs(t *testing.T) {
	m, reg := newMainForTest(t)

	m.activeConn = "A"
	m.openNewQueryTabWithRequest(manager.OpenQueryTabRequest{TabID: "a-tab", Name: "qa", Query: "select 1"})
	require.True(t, reg.Exists("a-tab"))

	m.DropWorkspace("A")
	assert.False(t, reg.Exists("a-tab"), "dropping the active workspace must unregister its tabs")
	assert.Empty(t, m.activeConn)
}

// Dropping a non-active workspace must not touch the live (active) workspace's
// registered tabs.
func TestMain_DropWorkspace_NonActiveLeavesActiveTabs(t *testing.T) {
	m, reg := newMainForTest(t)

	m.activeConn = "A"
	m.openNewQueryTabWithRequest(manager.OpenQueryTabRequest{TabID: "a-tab", Name: "qa", Query: "select 1"})
	require.True(t, reg.Exists("a-tab"))

	m.DropWorkspace("B") // some other, non-active connection
	assert.True(t, reg.Exists("a-tab"), "dropping a non-active workspace must not unregister active tabs")
	assert.Equal(t, "A", m.activeConn)
}
