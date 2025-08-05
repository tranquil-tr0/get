package output

import "context"

// Output defines the interface for printing different types of messages
// with optional formatting arguments.

type Output interface {
	// PrintStatus prints a message explaining what the program is doing with optional arguments
	PrintStatus(msg string, args ...any)
	// PrintSuccess prints a success message with optional arguments
	PrintSuccess(msg string, args ...any)
	// PrintError prints an error message with optional arguments
	PrintError(msg string, args ...any)
	// PrintInfo prints an informational message with optional arguments
	PrintInfo(msg string, args ...any)

	// PromptAssetIndexSelection presents asset name lists and returns the selected index
	PromptAssetIndexSelection(ctx context.Context, debNames, binaryNames, otherNames []string) (idx int, err error)

	// PromptElevatedCommand executes a command with elevated privileges and returns the output and error
	PromptElevatedCommand(prompt string, command string, args ...string) ([]byte, error)
}
