package output

import (
	"fmt"
	"os"
)

// Red returns the text in red color
func Red(text string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}

// PrintError prints the error message in red color to stderr
func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s\n", Red(msg))
}
