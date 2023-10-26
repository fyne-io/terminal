package terminal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

var apcHandlers = map[string]func(*Terminal, string){
	"set printer /windows-queue:": setWindowsQueue,
	"set printer:_file:":          setFilePrinting,
	"set printer:_editor:":        setEditorPrinting,
}

func (t *Terminal) handleAPC(code string) {
	for apcCommand, handler := range apcHandlers {
		if strings.HasPrefix(code, apcCommand) {
			// Extract the argument from the code
			arg := code[len(apcCommand):]
			// Invoke the corresponding handler function
			handler(t, arg)
			return
		}
	}

	if t.debug {
		// Handle other APC sequences or log the received APC code
		log.Println("Unrecognised APC", code)
	}

}

func setWindowsQueue(t *Terminal, arg string) {
	// Implement the action for setting the Windows queue
	log.Println("Setting Windows queue to", arg)
}

func setFilePrinting(t *Terminal, arg string) {
	t.printer = PrinterFunc(func(data []byte) {
		// Write data to the file
		err := writeToFile(arg, data)
		if err != nil && t.debug {
			log.Println("Error writing to file", err)
			return
		}
	})
}

// setEditorPrinting sets the printer to the printing editor.
func setEditorPrinting(t *Terminal, arg string) {
	t.printer = PrinterFunc(func(data []byte) {
		tempDir, err := os.MkdirTemp("", "print-data")
		if err != nil && t.debug {
			log.Println("Error creating temporary directory", err)
		}
		_, file := path.Split(arg)
		filename := path.Join(tempDir, file)
		// Write data to the file
		err = writeToFile(filename, data)
		if err != nil && t.debug {
			log.Println("Error writing to file", err)
			return
		}
		// Open the default application for the file
		err = openDefaultApp(filename)
		if err != nil && t.debug {
			log.Println("Error opening default application", err)
		}
	})
}

// Function to write data to a file
func writeToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// Function to open the default application for a file
func openDefaultApp(filename string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", filename).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", filename).Run()
	case "darwin":
		return exec.Command("open", filename).Run()
	case "android":
		// Assuming you have a shell on the Android device
		return exec.Command("am", "start", "-a", "android.intent.action.VIEW", "-d", "file://"+filename).Run()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
