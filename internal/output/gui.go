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
	fmt.Printf("GUI Status: "+msg+"\n", args...)
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

// PromptAssetIndexSelection presents asset name lists to the user in a dialog and returns the selected index.
func (o *GUIOutput) PromptAssetIndexSelection(debNames, binaryNames, otherNames []string) (int, error) {
	var allNames []string
	var options []string

	for _, name := range debNames {
		allNames = append(allNames, name)
		options = append(options, "[deb] "+name)
	}
	for _, name := range binaryNames {
		allNames = append(allNames, name)
		options = append(options, "[bin] "+name)
	}
	for _, name := range otherNames {
		allNames = append(allNames, name)
		options = append(options, "[other] "+name)
	}

	if len(allNames) == 0 {
		o.PrintError("No installable assets found.")
		return -1, nil
	}

	var ok bool
	item := qt.QInputDialog_GetItem4(o.window.QWidget, "Select Asset", "Choose an asset to install:", options, 0, false, &ok)
	if !ok {
		// User cancelled the dialog
		return -1, nil
	}
	for i, name := range options {
		if name == item {
			return i, nil
		}
	}
	return -1, nil
}
