package output

import (
	"context"
	"fmt"
	"os/exec"

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

// PrintStatus prints an action message in the main window's status bar.
func (o *GUIOutput) PrintStatus(msg string, args ...any) {
	if o.window == nil {
		return
	}
	statusBar := o.window.StatusBar()
	if statusBar == nil {
		statusBar = qt.NewQStatusBar(o.window.QWidget)
		o.window.SetStatusBar(statusBar)
	}
	out := fmt.Sprintf(msg, args...)
	statusBar.ShowMessage(out)
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
func (o *GUIOutput) PromptAssetIndexSelection(ctx context.Context, debNames, binaryNames, otherNames []string) (idx int, err error) {
	var options []string

	for _, name := range debNames {
		options = append(options, "[deb] "+name)
	}
	for _, name := range binaryNames {
		options = append(options, "[bin] "+name)
	}

	if len(options) == 0 {
		o.PrintError("No release packages found.")
		return -1, nil
	}

	var ok bool
	item := qt.QInputDialog_GetItem4(o.window.QWidget, "Select Asset", "Choose an asset to install:", options, 0, false, &ok)
	if !ok {
		// User cancelled the dialog
		return -1, context.Canceled
	}
	for i, name := range options {
		if name == item {
			return i, nil
		}
	}
	return -1, fmt.Errorf("selected item not found in options")
}

func (o *GUIOutput) PromptElevatedCommand(prompt string, command string, args ...string) ([]byte, error) {
	cmd := exec.Command("pkexec", append([]string{command}, args...)...)
	return cmd.CombinedOutput()
}

func (o *GUIOutput) PromptYesNo(msg string) (bool, error) {
	msgBox := qt.NewQMessageBox(o.window.QWidget)
	msgBox.SetIcon(qt.QMessageBox__Question)
	msgBox.SetWindowTitle("Question")
	msgBox.SetText(msg)
	msgBox.SetStandardButtons(qt.QMessageBox__Yes | qt.QMessageBox__No)
	result := msgBox.Exec()

	if result == int(qt.QMessageBox__Yes) {
		return true, nil
	} else if result == int(qt.QMessageBox__No) {
		return false, nil
	}
	return false, fmt.Errorf("unexpected dialog result: %d", result)
}
