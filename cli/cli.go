package cli

import (
	"bufio"
	"fmt"
	"os"
	// 	"strconv"
	//	"time"
)

// RawInput interactively prompts user
// and returns result.
func RawInput(prompt string) string {
	fmt.Printf("%s", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	selection := scanner.Text()
	return selection
}

func init() {
	fmt.Printf("Initializing cli package\n")
}
