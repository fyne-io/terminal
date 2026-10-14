package main

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"

	"github.com/stretchr/testify/assert"
)

// TestTabResize verifies the window grows when the tab bar first appears
// (going from 1 to 2 tabs), stays the same size for any further additions,
// and only shrinks back when the second-to-last tab is removed (back to 1).
func TestTabResize(t *testing.T) {
	a := test.NewApp()
	defer test.NewApp() // reset the global app after the test

	w, tabs, updateView := buildTerminalWindow(a, false, false)
	th := newTermTheme()
	dirs := make(map[*container.TabItem]string)

	addTab := func() {
		item := newTab(tabs, dirs, updateView, false, th, w, a, false)
		tabs.Append(item)
		tabs.Select(item)
		updateView(true)
	}
	removeLastTab := func() {
		tabs.Remove(tabs.Items[len(tabs.Items)-1])
		updateView(false)
	}

	// Single tab: no tab bar shown yet.
	single := w.Canvas().Size().Height
	assert.Equal(t, 1, len(tabs.Items))

	// Adding the second tab makes the tab bar appear and expands the window.
	addTab()
	twoTabs := w.Canvas().Size().Height
	assert.Equal(t, 2, len(tabs.Items))
	assert.Greater(t, twoTabs, single)

	// Adding a third tab must not change the window size.
	addTab()
	assert.Equal(t, 3, len(tabs.Items))
	assert.Equal(t, twoTabs, w.Canvas().Size().Height)

	// Removing a tab while more than two remain keeps the size the same.
	removeLastTab()
	assert.Equal(t, 2, len(tabs.Items))
	assert.Equal(t, twoTabs, w.Canvas().Size().Height)

	// Removing the second-to-last tab hides the tab bar and shrinks back.
	removeLastTab()
	assert.Equal(t, 1, len(tabs.Items))
	assert.Equal(t, single, w.Canvas().Size().Height)
}

// TestTabStartDir verifies a new tab opens in the working directory of the
// selected tab, ignoring a directory that has since been removed.
func TestTabStartDir(t *testing.T) {
	a := test.NewApp()
	defer test.NewApp() // reset the global app after the test

	w, tabs, updateView := buildTerminalWindow(a, false, false)
	th := newTermTheme()
	dirs := make(map[*container.TabItem]string)

	dir := t.TempDir()
	dirs[tabs.Selected()] = dir
	item := newTab(tabs, dirs, updateView, false, th, w, a, false)
	assert.Equal(t, dir, dirs[item])
	assert.Equal(t, dir, currentDir(tabs, dirs))

	missing := filepath.Join(dir, "missing")
	dirs[tabs.Selected()] = missing
	item = newTab(tabs, dirs, updateView, false, th, w, a, false)
	assert.NotEqual(t, missing, dirs[item])
	assert.Equal(t, "", currentDir(tabs, dirs))

	file := filepath.Join(dir, "file")
	assert.NoError(t, os.WriteFile(file, []byte{}, 0o600))
	dirs[tabs.Selected()] = file
	assert.Equal(t, "", currentDir(tabs, dirs))
}

// TestTabClose verifies that closing a tab with its close button tidies up
// in the same way as the shell exiting, hiding the tab bar for a single tab.
func TestTabClose(t *testing.T) {
	a := test.NewApp()
	defer test.NewApp() // reset the global app after the test

	w, tabs, updateView := buildTerminalWindow(a, false, false)
	th := newTermTheme()
	dirs := make(map[*container.TabItem]string)
	single := w.Canvas().Size().Height

	item := newTab(tabs, dirs, updateView, false, th, w, a, false)
	item.Text = "Closing"
	tabs.Append(item)
	tabs.Select(item)
	updateView(true)
	assert.Equal(t, termTitle()+": Closing", w.Title())
	assert.Greater(t, w.Canvas().Size().Height, single)

	// this is what the close button of a tab does
	tabs.Remove(item)
	tabs.OnClosed(item)

	assert.Equal(t, 1, len(tabs.Items))
	assert.Equal(t, single, w.Canvas().Size().Height)
	assert.Equal(t, termTitle(), w.Title())
	assert.Equal(t, findTerminal(tabs.Items[0]), w.Canvas().Focused())
}
