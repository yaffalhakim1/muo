package muo

import (
	"github.com/egoist/mygo"
)

// The menu bar. Every command the window offers is here as well, because that
// is where people look for a command they cannot find, and a desktop app that
// keeps one only in a toolbar hides it from keyboard users
// (the-menu-bar.md › Menu bar menus: "Provide app-specific menus for custom
// commands. People look in the menu bar when searching for app-specific
// commands").

func (r *Reader) InstallMenu() {
	file := &mygo.MenuItem{
		Label: "File", Type: mygo.MenuItemSubmenu, Submenu: []*mygo.MenuItem{
			{Label: "Open a File…", Accelerator: "CmdOrCtrl+O", Click: func(*mygo.MenuItem, *mygo.Window) { r.Open() }},
			// Open Recent is a system role: it lists the same files the empty
			// page does, and choosing one reaches the app through OnOpenFile.
			{Role: mygo.RoleOpenRecent},
			{Type: mygo.MenuItemSeparator},
			{Label: "Reload Document", Accelerator: "CmdOrCtrl+R", Click: func(*mygo.MenuItem, *mygo.Window) { r.Reload() }},
			{Type: mygo.MenuItemSeparator},
			{Label: "Close Document", Accelerator: "CmdOrCtrl+W", Click: func(*mygo.MenuItem, *mygo.Window) {
				if r.active >= 0 {
					r.closeDocument(r.active)
				}
			}},
			{Type: mygo.MenuItemSeparator},
			{Role: mygo.RoleClose},
		},
	}
	edit := &mygo.MenuItem{
		Label: "Edit", Type: mygo.MenuItemSubmenu, Submenu: []*mygo.MenuItem{
			{Role: mygo.RoleUndo},
			{Role: mygo.RoleRedo},
			{Type: mygo.MenuItemSeparator},
			{Role: mygo.RoleCut},
			{Role: mygo.RoleCopy},
			{Role: mygo.RolePaste},
			{Role: mygo.RoleSelectAll},
			{Type: mygo.MenuItemSeparator},
			{Label: "Find…", Accelerator: "CmdOrCtrl+F", Click: func(*mygo.MenuItem, *mygo.Window) {
				r.finding = true
			}},
		},
	}
	// The two panes come first in a View menu, as Apple's own apps have them, and
	// they are here as well as in the title row because that is where people look
	// for a command they cannot find (the-menu-bar.md › View menu). Each is one
	// item whose label says what it will do, so the menu is rebuilt when a pane
	// changes — the alternative, two items sharing one shortcut, is a shortcut
	// that only works in one direction.
	libraryItem := &mygo.MenuItem{
		Label:       "Show Library",
		Accelerator: "CmdOrCtrl+Ctrl+L",
		Click: func(*mygo.MenuItem, *mygo.Window) {
			r.showLibrary = true
			r.narrowLibrary = true
			r.refreshMenu()
		},
	}
	if r.showLibrary {
		libraryItem.Label = "Hide Library"
		libraryItem.Click = func(*mygo.MenuItem, *mygo.Window) {
			r.showLibrary = false
			r.narrowLibrary = false
			r.refreshMenu()
		}
	}
	outlineItem := &mygo.MenuItem{
		Label:       "Show Page Outline",
		Accelerator: "CmdOrCtrl+Ctrl+O",
		Click: func(*mygo.MenuItem, *mygo.Window) {
			r.showOutline = true
			r.narrowOutline = true
			r.refreshMenu()
		},
	}
	if r.showOutline {
		outlineItem.Label = "Hide Page Outline"
		outlineItem.Click = func(*mygo.MenuItem, *mygo.Window) {
			r.showOutline = false
			r.narrowOutline = false
			r.refreshMenu()
		}
	}
	view := &mygo.MenuItem{
		Label: "View", Type: mygo.MenuItemSubmenu, Submenu: []*mygo.MenuItem{
			libraryItem,
			outlineItem,
			{Type: mygo.MenuItemSeparator},
			{Label: "Reading Settings…", Accelerator: "CmdOrCtrl+,", Click: func(*mygo.MenuItem, *mygo.Window) {
				r.settings = true
			}},
			{Type: mygo.MenuItemSeparator},
			{Label: "Next Page", Accelerator: "Alt+Down", Click: func(*mygo.MenuItem, *mygo.Window) {
				r.goToPage(r.viewing + 1)
			}},
			{Label: "Previous Page", Accelerator: "Alt+Up", Click: func(*mygo.MenuItem, *mygo.Window) {
				r.goToPage(r.viewing - 1)
			}},
			{Type: mygo.MenuItemSeparator},
			{Label: "Larger Type", Accelerator: "CmdOrCtrl+=", Click: func(*mygo.MenuItem, *mygo.Window) {
				r.stepType(1.05)
			}},
			{Label: "Smaller Type", Accelerator: "CmdOrCtrl+-", Click: func(*mygo.MenuItem, *mygo.Window) {
				r.stepType(1 / 1.05)
			}},
			{Label: "Follow the Model", Accelerator: "CmdOrCtrl+0", Click: func(*mygo.MenuItem, *mygo.Window) {
				r.manual = 0
				if r.display != nil {
					r.display.FontPx = 0
				}
			}},
			{Type: mygo.MenuItemSeparator},
			{Role: mygo.RoleToggleFullScreen},
		},
	}
	window := &mygo.MenuItem{
		Label: "Window", Type: mygo.MenuItemSubmenu, Submenu: []*mygo.MenuItem{
			{Role: mygo.RoleMinimize},
			{Role: mygo.RoleZoom},
		},
	}
	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{file, edit, view, window}))
}

// stepType changes the manual type size by a factor, from whatever size is in
// effect, so that a step from the model is smooth.
func (r *Reader) stepType(factor float32) {
	if r.display == nil {
		return
	}
	base := r.typ.FontPx
	if r.display.FontPx > 0 {
		base = r.display.FontPx
	}
	next := base * factor
	if next < 11 {
		next = 11
	}
	if next > 64 {
		next = 64
	}
	r.manual = next
	r.display.FontPx = next
}

