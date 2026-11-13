package mcp

import (
	"fmt"

	"github.com/kopecmaciej/vi-sql/internal/database"
)

// ConnInfo describes an open connection, returned by the list_connections tool.
type ConnInfo struct {
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Database string `json:"database,omitempty"`
	Active   bool   `json:"active"`
}

// ConnectionResolver gives the MCP server live access to the set of open
// database connections, so individual tool calls can target a specific one.
type ConnectionResolver interface {
	// ResolveDriver returns the driver for the named connection. An empty name
	// resolves to the active connection. ok is false when no matching open
	// connection exists.
	ResolveDriver(name string) (driver database.Driver, ok bool)
	// ListConnections returns metadata for every open connection.
	ListConnections() []ConnInfo
}

// resolveConnName returns the concrete name a tool call targeted. An empty name
// resolves to the active connection's name; it stays empty if none is open.
func (s *Server) resolveConnName(name string) string {
	if name != "" || s.connections == nil {
		return name
	}
	for _, c := range s.connections.ListConnections() {
		if c.Active {
			return c.Name
		}
	}
	return name
}

// driverFor resolves the driver a tool should run against. An empty name uses
// the active connection; an unknown name yields an error listing what's open.
func (s *Server) driverFor(name string) (database.Driver, error) {
	if s.connections == nil {
		return nil, fmt.Errorf("no database connection available")
	}
	if driver, ok := s.connections.ResolveDriver(name); ok {
		return driver, nil
	}
	if name == "" {
		return nil, fmt.Errorf("no active database connection")
	}
	open := s.connections.ListConnections()
	names := make([]string, len(open))
	for i, c := range open {
		names[i] = c.Name
	}
	return nil, fmt.Errorf("connection %q is not open; open connections: %v", name, names)
}
