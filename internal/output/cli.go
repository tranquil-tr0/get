package output

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
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
func (o *CLIOutput) PromptAssetIndexSelection(ctx context.Context, debNames, binaryNames, otherNames []string) (idx int, err error) {
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
	// showOther variable removed (was unused)

	otherShown := false
	if len(otherNames) > 0 {
		fmt.Printf("  [s] Show other assets (%d)\n", len(otherNames))
	}

	if len(allNames) == 0 && len(otherNames) == 0 {
		fmt.Println("No installable assets found.")
		return -1, fmt.Errorf("no installable assets")
	}

	for {
		// Only show the 's' option if other assets haven't been shown yet and there are other assets
		prompt := "\nSelect an option by entering a number"
		if len(otherNames) > 0 && !otherShown {
			prompt += ", 's' to show other assets"
		}
		if len(otherNames) > 0 && otherShown {
			prompt += " to manually define it as a binary and install it as such"
		}
		prompt += ", or 'c' to cancel: "
		fmt.Print(prompt)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if input == "c" || input == "C" || input == "" {
			return -1, context.Canceled
		}
		if (input == "s" || input == "S") && len(otherNames) > 0 && !otherShown {
			// Show other assets
			for _, name := range otherNames {
				allNames = append(allNames, name)
				fmt.Printf("  [%d] [other] %s\n", len(allNames), name)
			}
			otherShown = true
			continue
		}
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(allNames) {
			fmt.Printf("Invalid selection: %s\n", input)
			continue
		}
		return choice - 1, nil
	}
}

func (o *CLIOutput) PromptElevatedCommand(prompt string, command string, args ...string) ([]byte, error) {
	prompt = "[get] " + prompt
	cmd := exec.Command("sudo", append([]string{"-p", prompt, command}, args...)...)
	return cmd.CombinedOutput()
}

func (o *CLIOutput) PromptYesNo(msg string) (bool, error) {
	fmt.Printf("%s [Y/n]: ", msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))
	switch input {
	case "", "y":
		return true, nil
	case "n":
		return false, nil
	default:
		fmt.Printf("Invalid input: %s. Please enter 'y' or 'n'.\n", input)
		return o.PromptYesNo(msg) // Retry the prompt
	}
}
