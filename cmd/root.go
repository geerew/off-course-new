package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "off-course",
	Short: "Off Course",
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Execute is the entry point for the CLI
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// questionPlain asks a question, waits for plain text input and returns the answer
func questionPlain(question string) string {
	c := color.New(color.Bold, color.FgGreen)
	c.Printf(">> %s: ", question)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)

	return answer
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// questionPlain asks a question, waits for hidden text input and returns the answer
func questionPassword(question string) string {
	c := color.New(color.Bold, color.FgGreen)
	c.Printf(">> %s: ", question)

	answerBytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
	answer := string(answerBytes)
	answer = strings.TrimSpace(answer)

	c.Println()

	return answer
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// errorMessage prints an error message
func errorMessage(message string, a ...any) {
	c := color.New(color.FgRed)
	c.Printf(message+"\n", a...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// successMessage prints a success message
func successMessage(message string, a ...any) {
	c := color.New(color.FgYellow)
	c.Printf(message+"\n", a...)
}
