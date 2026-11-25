package modal

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-sql/internal/manager"
	"github.com/kopecmaciej/vi-sql/internal/tui/core"
)

const (
	ConnectionSwitcherId = "ConnectionSwitcher"
	ReadOnlyMarker       = "⊘"
)

// ConnEntry is a single open connection shown in the switcher.
type ConnEntry struct {
	Name     string
	Detail   string
	Active   bool
	ReadOnly bool
}

// ConnectionSwitcher is a list of the currently open connections. Enter
// switches to the highlighted connection; the close-tab key closes it.
type ConnectionSwitcher struct {
	*core.BaseElement
	*core.Flex

	list     *core.Table
	entries  []ConnEntry
	onSwitch func(name string)
	onClose  func(name string)
}

func NewConnectionSwitcher() *ConnectionSwitcher {
	c := &ConnectionSwitcher{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		list:        core.NewTable(),
	}
	c.SetIdentifier(ConnectionSwitcherId)
	c.SetAfterInitFunc(c.init)
	return c
}

func (c *ConnectionSwitcher) init() error {
	c.setLayout()
	c.setStyle()
	c.setKeybindings()
	c.handleEvents()
	return nil
}

func (c *ConnectionSwitcher) handleEvents() {
	go c.HandleEvents(ConnectionSwitcherId, func(event manager.EventMsg) {
		if event.Message.Type == manager.StyleChanged {
			c.setStyle()
		}
	})
}

func (c *ConnectionSwitcher) setLayout() {
	c.Flex.SetDirection(tview.FlexRow)
	c.Flex.SetBorder(true)
	c.Flex.SetTitle(" Open connections ")
	c.Flex.AddItem(c.list, 0, 1, true)

	c.list.SetSelectable(true, false)
	c.list.SetBorderPadding(0, 0, 1, 1)
}

func (c *ConnectionSwitcher) setStyle() {
	styles := c.App.GetStyles()
	c.Flex.SetStyle(styles)
	c.list.SetStyle(styles)
}

func (c *ConnectionSwitcher) setKeybindings() {
	k := c.App.GetKeys()
	c.list.SetInputCapture(k.WrapInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyDown || k.Match(k.Navigation.FocusDown, event):
			c.moveSelection(1)
			return nil
		case event.Key() == tcell.KeyUp || k.Match(k.Navigation.FocusUp, event):
			c.moveSelection(-1)
			return nil
		case event.Key() == tcell.KeyEnter:
			c.switchSelected()
			return nil
		case k.Match(k.Main.CloseTab, event):
			c.closeSelected()
			return nil
		case event.Key() == tcell.KeyEscape || k.Match(k.Common.Close, event):
			c.App.Pages.RemoveModalPage(ConnectionSwitcherId)
			return nil
		}
		return event
	}))
}

func (c *ConnectionSwitcher) moveSelection(delta int) {
	if len(c.entries) == 0 {
		return
	}
	row, _ := c.list.GetSelection()
	next := row + delta
	if next < 0 || next >= len(c.entries) {
		return
	}
	c.list.Select(next, 0)
}

func (c *ConnectionSwitcher) selectedName() string {
	row, _ := c.list.GetSelection()
	if row < 0 || row >= len(c.entries) {
		return ""
	}
	return c.entries[row].Name
}

func (c *ConnectionSwitcher) switchSelected() {
	name := c.selectedName()
	c.App.Pages.RemoveModalPage(ConnectionSwitcherId)
	if name != "" && c.onSwitch != nil {
		c.onSwitch(name)
	}
}

func (c *ConnectionSwitcher) closeSelected() {
	name := c.selectedName()
	c.App.Pages.RemoveModalPage(ConnectionSwitcherId)
	if name != "" && c.onClose != nil {
		c.onClose(name)
	}
}

func (c *ConnectionSwitcher) renderList() {
	c.list.Clear()
	styles := c.App.GetStyles()
	for i, e := range c.entries {
		marker := "  "
		nameColor := styles.Global.TextColor.Color()
		if e.Active {
			marker = "● "
			nameColor = styles.Global.SecondaryTextColor.Color()
		}
		name := marker + e.Name
		if e.ReadOnly {
			name += fmt.Sprintf(" [%s]%s[-]", styles.Global.WarningColor, ReadOnlyMarker)
		}
		c.list.SetCell(i, 0, tview.NewTableCell(name).
			SetTextColor(nameColor))
		c.list.SetCell(i, 1, tview.NewTableCell(e.Detail+" ").
			SetTextColor(styles.Global.DimColor.Color()).
			SetAlign(tview.AlignRight).
			SetExpansion(1))
	}
}

// Open shows the switcher for the given open connections. onSwitch is invoked
// with the chosen connection name; onClose with the name to close.
func (c *ConnectionSwitcher) Open(entries []ConnEntry, onSwitch, onClose func(name string)) {
	c.entries = entries
	c.onSwitch = onSwitch
	c.onClose = onClose
	c.renderList()

	for i, e := range entries {
		if e.Active {
			c.list.Select(i, 0)
			break
		}
	}

	wrapper := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(c.Flex, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	c.App.Pages.ShowModal(ConnectionSwitcherId, wrapper, c.list, true, true)
}
