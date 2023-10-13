package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/fyne-io/terminal"
)

// No frills
// This is just an example of the bare minimum terminal. I wouldn't use this as a replacement for putty or iterm
// From here you can build themes and tabs, shortcuts.
// Now that you know the basics, take a look at fyneterm and see a more robust terminal example. 

func main() {
	a := app.New()
	w := a.NewWindow("Terminal")
	t := terminal.New()
	w.SetContent(t)
	w.Canvas().Focus(t)
	w.Resize(fyne.Size{Width: 800, Height: 600})
	w.SetPadded(false)
	go func() {
		err := t.RunLocalShell()
		if err != nil {
			fyne.LogError("Failure in terminal", err)
		}

		w.Close()
	}()
	w.ShowAndRun()
	fmt.Println("Done")
}
