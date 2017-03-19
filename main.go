package main

import (
	"fmt"
	"os"
)

// Version
const version = "v0.37.10-alpha"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
