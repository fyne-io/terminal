package main

import (
	"testing"

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

	addTab := func() {
		item := newTab(tabs, updateView, false, th, w, a, false)
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
