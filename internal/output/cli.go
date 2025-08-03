package output

// CLIOutput implements the Output interface for the command-line interface.

type CLIOutput struct{}

// NewCLIOutput creates a new CLIOutput.
func NewCLIOutput() *CLIOutput {
	return &CLIOutput{}
}

// PrintStatus prints an action message to the console.
func (o *CLIOutput) PrintStatus(msg string, args ...any) {
	PrintStatus(msg, args...)
}

// PrintSuccess prints a success message to the console.
func (o *CLIOutput) PrintSuccess(msg string, args ...any) {
	PrintSuccess(msg, args...)
}

// PrintError prints an error message to the console.
func (o *CLIOutput) PrintError(msg string, args ...any) {
	PrintError(msg, args...)
}

// PrintInfo prints an informational message to the console.
func (o *CLIOutput) PrintInfo(msg string, args ...any) {
	PrintInfo(msg, args...)
}
