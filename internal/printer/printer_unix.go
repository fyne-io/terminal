//go:build !windows

package printer

// PrintPostScriptFile writes the filedata to the printer.
func PrintPostScriptFile(printerName string, filedata []byte) error {
	// TODO implement me.
	return nil
}
