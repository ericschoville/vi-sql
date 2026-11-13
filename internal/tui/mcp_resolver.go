package tui

import (
	"github.com/kopecmaciej/vi-sql/internal/database"
	visqlmcp "github.com/kopecmaciej/vi-sql/internal/mcp"
	"github.com/kopecmaciej/vi-sql/internal/tui/core"
)

// mcpConnResolver adapts the app's open-connection registry to the MCP server's
// ConnectionResolver interface, letting MCP tools target any open connection. It
// resolves against live app state, so the server need not restart on a switch.
type mcpConnResolver struct {
	app *core.App
}

func (r mcpConnResolver) ResolveDriver(name string) (database.Driver, bool) {
	if name == "" {
		name = r.app.GetConfig().CurrentConnection
	}
	oc, ok := r.app.GetOpenConn(name)
	if !ok {
		return nil, false
	}
	return oc.Driver, true
}

func (r mcpConnResolver) ListConnections() []visqlmcp.ConnInfo {
	active := r.app.GetConfig().CurrentConnection
	conns := r.app.OpenConns()
	out := make([]visqlmcp.ConnInfo, 0, len(conns))
	for _, oc := range conns {
		out = append(out, visqlmcp.ConnInfo{
			Name:     oc.Config.Name,
			Driver:   oc.Config.GetDriver(),
			Database: oc.Config.Database,
			Active:   oc.Config.Name == active,
		})
	}
	return out
}
