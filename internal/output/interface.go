package output

import "context"

// Output defines the interface for printing different types of messages
// with optional formatting arguments.

type Output interface {
	// PrintStatus shows a message explaining what the program is doing
	PrintStatus(msg string, args ...any)
	// PrintSuccess shows a success message, such as after an action completes
	PrintSuccess(msg string, args ...any)
	// PrintError shows an error message
	PrintError(msg string, args ...any)
	// PrintInfo notifies the user of some information, such as the results of an action or other important information
	PrintInfo(msg string, args ...any)

	// PromptAssetIndexSelection presents asset name lists and returns the selected index
	PromptAssetIndexSelection(ctx context.Context, debNames, binaryNames, otherNames []string) (idx int, err error)

	// PromptElevatedCommand executes a command with elevated privileges and returns the output and error
	PromptElevatedCommand(prompt string, command string, args ...string) ([]byte, error)

	// PromptYesNo presents a yes/no question to the user and returns their answer
	PromptYesNo(msg string) (bool, error)
}
