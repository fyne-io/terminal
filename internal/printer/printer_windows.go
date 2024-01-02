//go:build windows

package printer

import (
	"fmt"

	"github.com/alexbrainman/printer"
)

// PrintPostScriptFile writes the filedata to the printer.
func PrintPostScriptFile(printerName string, filedata []byte) error {

	if printerName == "" {
		var err error
		// Get the default printer.
		printerName, err = printer.Default()
		if err != nil {
			return err
		}
	}
	// Open the printer.
	p, err := printer.Open(printerName)
	if err != nil {
		return fmt.Errorf("error opening printer: %w", err)
	}
	defer p.Close()

	// Start a raw document
	err = p.StartRawDocument("PrintJob")
	if err != nil {
		return err
	}
	defer p.EndDocument()

	// Write the data.
	_, err = p.Write(filedata)
	if err != nil {
		return fmt.Errorf("error writing data to printer: %w", err)
	}

	return nil
}
