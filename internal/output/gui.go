package output

import (
	"fmt"

	qt "github.com/mappu/miqt/qt6"
)

// GUIOutput implements the Output interface for the GUI.

type GUIOutput struct {
	window *qt.QMainWindow
}

// NewGUIOutput creates a new GUIOutput.
func NewGUIOutput(window *qt.QMainWindow) *GUIOutput {
	return &GUIOutput{window: window}
}

// PrintStatus prints an action message in a dialog.
func (o *GUIOutput) PrintStatus(msg string, args ...any) {
	// TODO: Implement status bar for action messages
	// Temporarily using console output for debugging
	fmt.Printf("GUI Action: "+msg+"\n", args...)
}

// PrintSuccess prints a success message in a dialog.
func (o *GUIOutput) PrintSuccess(msg string, args ...any) {
	msgBox := qt.NewQMessageBox(o.window.QWidget)
	msgBox.SetIcon(qt.QMessageBox__Information)
	msgBox.SetWindowTitle("Success")
	msgBox.SetText(fmt.Sprintf(msg, args...))
	msgBox.Exec()
}

// PrintError prints an error message in a dialog.
func (o *GUIOutput) PrintError(msg string, args ...any) {
	msgBox := qt.NewQMessageBox(o.window.QWidget)
	msgBox.SetIcon(qt.QMessageBox__Critical)
	msgBox.SetWindowTitle("Error")
	msgBox.SetText(fmt.Sprintf(msg, args...))
	msgBox.Exec()
}

// PrintInfo prints an informational message in a dialog.
func (o *GUIOutput) PrintInfo(msg string, args ...any) {
	msgBox := qt.NewQMessageBox(o.window.QWidget)
	msgBox.SetIcon(qt.QMessageBox__Information)
	msgBox.SetWindowTitle("Info")
	msgBox.SetText(fmt.Sprintf(msg, args...))
	msgBox.Exec()
}
