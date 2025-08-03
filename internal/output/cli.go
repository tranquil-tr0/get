package output

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CLIOutput implements the Output interface for the command-line interface.

type CLIOutput struct{}

// NewCLIOutput creates a new CLIOutput.
func NewCLIOutput() *CLIOutput {
	return &CLIOutput{}
}

// PrintStatus prints an action message to the console.
func (c *CLIOutput) PrintStatus(msg string, args ...any) {
	out := fmt.Sprintf(msg, args...)
	fmt.Printf("\033[33m%s\033[0m\n", out)
}

// PrintSuccess prints a success message to the console.
func (c *CLIOutput) PrintSuccess(msg string, args ...any) {
	out := fmt.Sprintf(msg, args...)
	// Bold green
	fmt.Printf("\033[1;32m%s\033[0m\n", out)
}

// PrintError prints an error message to the console stderr.
func (c *CLIOutput) PrintError(msg string, args ...any) {
	out := fmt.Sprintf(msg, args...)
	fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", out)
}

// PrintInfo prints an informational message to the console.
func (c *CLIOutput) PrintInfo(msg string, args ...any) {
	out := fmt.Sprintf(msg, args...)
	fmt.Printf("%s\n", out)
}

// PromptAssetIndexSelection presents asset name lists to the user and returns the selected index.
func (o *CLIOutput) PromptAssetIndexSelection(debNames, binaryNames, otherNames []string) (int, error) {
	var allNames []string

	fmt.Println("\nAvailable assets:")
	for _, name := range debNames {
		allNames = append(allNames, name)
		fmt.Printf("  [%d] [deb] %s\n", len(allNames), name)
	}
	for _, name := range binaryNames {
		allNames = append(allNames, name)
		fmt.Printf("  [%d] [bin] %s\n", len(allNames), name)
	}
	for _, name := range otherNames {
		allNames = append(allNames, name)
		fmt.Printf("  [%d] [other] %s\n", len(allNames), name)
	}

	if len(allNames) == 0 {
		fmt.Println("No installable assets found.")
		return -1, fmt.Errorf("no installable assets")
	}

	fmt.Print("\nSelect an option (number): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(allNames) {
		return -1, fmt.Errorf("invalid selection: %s", input)
	}
	return choice - 1, nil
}
