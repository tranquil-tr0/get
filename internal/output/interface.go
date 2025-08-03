package output

// Output defines the interface for printing different types of messages
// with optional formatting arguments.
type Output interface {
	// PrintAction prints an action message with optional arguments
	PrintAction(msg string, args ...any)
	// PrintSuccess prints a success message with optional arguments
	PrintSuccess(msg string, args ...any)
	// PrintError prints an error message with optional arguments
	PrintError(msg string, args ...any)
	// PrintInfo prints an informational message with optional arguments
	PrintInfo(msg string, args ...any)
}
