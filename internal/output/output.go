package output

import (
	"fmt"
	"os"
	"time"
)

// Global verbose state
var isVerbose bool

// SetVerbose sets the global verbose state
func SetVerbose(verbose bool) {
	isVerbose = verbose
}

// IsVerbose returns the current verbose state
func IsVerbose() bool {
	return isVerbose
}

// Bold returns the text in bold format
func Bold(text string) string {
	return fmt.Sprintf("\x1b[1m%s\x1b[0m", text)
}

// Red returns the text in red color
func Red(text string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}

// Green returns the text in green color
func Green(text string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", text)
}

// Yellow returns the text in yellow color
func Yellow(text string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", text)
}

// PrintTitle prints the text in bold format
func PrintTitle(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(Bold(msg))
}

// PrintError prints the error message in red color to stderr
func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s\n", Red(msg))
}

// PrintGreen prints the message in green color
func PrintGreen(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s\n", Green(msg))
}

// PrintYellow prints the message in yellow color
func PrintYellow(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s\n", Yellow(msg))
}

// PrintStatus calls PrintYellow
func PrintStatus(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	PrintYellow("%s", msg)
}

// PrintWarn prints the message in yellow color and bold format
func PrintWarn(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s\n", Bold(Yellow(msg)))
}

// PrintSuccess prints the message in green color and bold format
func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s\n", Bold(Green(msg)))
}

// PrintInfo prints the message in normal format
func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s\n", msg)
}

// Verbose logging functions that only print when verbose mode is enabled

// PrintVerbose prints verbose information in a dimmed format
func PrintVerbose(format string, a ...interface{}) {
	if !isVerbose {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\033[2m[VERBOSE] %s\033[0m\n", msg)
}

// PrintVerboseAction prints verbose action information with timestamp
func PrintVerboseAction(format string, a ...interface{}) {
	if !isVerbose {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\033[2m[%s] %s\033[0m\n", timestamp, msg)
}

// PrintVerboseStart prints when an action is starting
func PrintVerboseStart(action string, details ...interface{}) {
	if !isVerbose {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	if len(details) > 0 {
		fmt.Printf("\033[2m[%s] Starting: %s (%v)\033[0m\n", timestamp, action, details[0])
	} else {
		fmt.Printf("\033[2m[%s] Starting: %s\033[0m\n", timestamp, action)
	}
}

// PrintVerboseComplete prints when an action is completed
func PrintVerboseComplete(action string, details ...interface{}) {
	if !isVerbose {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	if len(details) > 0 {
		fmt.Printf("\033[2m[%s] ✓ Completed: %s (%v)\033[0m\n", timestamp, action, details[0])
	} else {
		fmt.Printf("\033[2m[%s] ✓ Completed: %s\033[0m\n", timestamp, action)
	}
}

// PrintVerboseError prints verbose error information
func PrintVerboseError(action string, err error) {
	if !isVerbose {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("\033[2m[%s] ✗ Failed: %s - %v\033[0m\n", timestamp, action, err)
}

// PrintVerboseDebug prints debug information with extra details
func PrintVerboseDebug(category string, format string, a ...interface{}) {
	if !isVerbose {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\033[2m[%s] [DEBUG:%s] %s\033[0m\n", timestamp, category, msg)
}
