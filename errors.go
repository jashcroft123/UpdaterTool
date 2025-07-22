package main

import (
	"fmt"
	"os"
)

// Error handling in a single place, nothing fancy to do for this program
func check(err error) {
	if err != nil {
		fmt.Println("Error: %s", err.Error())
		os.Exit(1)
	}
}
