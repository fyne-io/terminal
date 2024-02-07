package notosansmono

import (
	// go embed.
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed NotoSansMono-Regular.ttf
var regular []byte

// Regular is the regular font resource.
var Regular = &fyne.StaticResource{
	StaticName:    "NotoSansMono-Regular.ttf",
	StaticContent: regular,
}

//go:embed NotoSansMono-Bold.ttf
var bold []byte

// Bold is the bold font resource.
var Bold = &fyne.StaticResource{
	StaticName:    "NotoSansMono-Bold.ttf",
	StaticContent: bold,
}
