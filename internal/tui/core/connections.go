package core

import (
	"context"

	"github.com/kopecmaciej/vi-sql/internal/config"
	"github.com/kopecmaciej/vi-sql/internal/database"
)

// OpenConn is a live database connection held open by the app. The registry
// keeps one per connection so switching between them doesn't tear down and
// re-establish the underlying driver.
type OpenConn struct {
	Config    *config.SQLConfig
	Driver    database.Driver
	Formatter database.ValueFormatter
}

func (a *App) ensureConnMap() {
	if a.openConns == nil {
		a.openConns = make(map[string]*OpenConn)
	}
}

// AddOpenConn registers an open connection, preserving open order. Re-adding an
// existing name replaces its entry without duplicating the order slot.
func (a *App) AddOpenConn(oc *OpenConn) {
	a.ensureConnMap()
	name := oc.Config.Name
	if _, exists := a.openConns[name]; !exists {
		a.connOrder = append(a.connOrder, name)
	}
	a.openConns[name] = oc
}

// GetOpenConn returns the open connection registered under name, if any.
func (a *App) GetOpenConn(name string) (*OpenConn, bool) {
	oc, ok := a.openConns[name]
	return oc, ok
}

// IsConnOpen reports whether a connection with the given name is open.
func (a *App) IsConnOpen(name string) bool {
	_, ok := a.openConns[name]
	return ok
}

// OpenConns returns the open connections in the order they were opened.
func (a *App) OpenConns() []*OpenConn {
	out := make([]*OpenConn, 0, len(a.connOrder))
	for _, name := range a.connOrder {
		if oc, ok := a.openConns[name]; ok {
			out = append(out, oc)
		}
	}
	return out
}

// CloseConn closes the driver for name and removes it from the registry. It is
// a no-op if the connection isn't open.
func (a *App) CloseConn(ctx context.Context, name string) error {
	oc, ok := a.openConns[name]
	if !ok {
		return nil
	}
	delete(a.openConns, name)
	for i, n := range a.connOrder {
		if n == name {
			a.connOrder = append(a.connOrder[:i], a.connOrder[i+1:]...)
			break
		}
	}
	if oc.Driver != nil {
		return oc.Driver.Close(ctx)
	}
	return nil
}

// CloseAllConns closes every open connection's driver and empties the registry.
func (a *App) CloseAllConns(ctx context.Context) {
	for _, oc := range a.openConns {
		if oc.Driver != nil {
			_ = oc.Driver.Close(ctx)
		}
	}
	a.openConns = nil
	a.connOrder = nil
}
