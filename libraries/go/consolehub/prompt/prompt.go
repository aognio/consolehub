package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"consolehub/libraries/go/consolehub/events"
	"consolehub/libraries/go/consolehub/queue"

	"golang.org/x/term"
)

// Prompt asks a text question with optional default value.
func Prompt(promptText, defaultVal string, q *queue.BoundedQueue) string {
	fmt.Print(promptText)
	if defaultVal != "" {
		fmt.Printf(" [%s]", defaultVal)
	}
	fmt.Print(": ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		input = defaultVal
	}

	if q != nil {
		q.Push(events.NewPromptEvent(fmt.Sprintf("p-%s", promptText), promptText, input, false))
	}

	return input
}

// SecretPrompt asks for a secret password/API key without echoing text.
func SecretPrompt(promptText string, q *queue.BoundedQueue) string {
	fmt.Printf("%s: ", promptText)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	if err != nil {
		return ""
	}
	input := strings.TrimSpace(string(bytePassword))

	if q != nil {
		q.Push(events.NewPromptEvent(fmt.Sprintf("secret-%s", promptText), promptText, "********", true))
	}

	return input
}

// Confirm asks a yes/no question returning boolean.
func Confirm(promptText string, defaultVal bool, q *queue.BoundedQueue) bool {
	hint := "y/N"
	if defaultVal {
		hint = "Y/n"
	}
	fmt.Printf("%s [%s]: ", promptText, hint)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	result := defaultVal
	if input == "y" || input == "yes" {
		result = true
	} else if input == "n" || input == "no" {
		result = false
	}

	respStr := "no"
	if result {
		respStr = "yes"
	}

	if q != nil {
		q.Push(events.NewPromptEvent(fmt.Sprintf("confirm-%s", promptText), promptText, respStr, false))
	}

	return result
}

// Choice prompts selecting one option from a list of options.
func Choice(promptText string, options []string, defaultChoice string, q *queue.BoundedQueue) string {
	fmt.Println(promptText)
	for i, opt := range options {
		marker := " "
		if opt == defaultChoice {
			marker = "*"
		}
		fmt.Printf("  %d) %s %s\n", i+1, opt, marker)
	}
	fmt.Printf("Select choice (1-%d) [%s]: ", len(options), defaultChoice)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	selected := defaultChoice
	if input != "" {
		for i, opt := range options {
			if input == fmt.Sprintf("%d", i+1) || strings.EqualFold(input, opt) {
				selected = opt
				break
			}
		}
	}

	if q != nil {
		q.Push(events.NewPromptEvent(fmt.Sprintf("choice-%s", promptText), promptText, selected, false))
	}

	return selected
}
