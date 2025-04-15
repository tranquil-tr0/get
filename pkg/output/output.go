package output

import (
	"fmt"
	"os"
)

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

// PrintAction calls PrintYellow
func PrintAction(format string, a ...interface{}) {
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

// PrintNormal prints the message in normal format
func PrintNormal(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s\n", msg)
}
