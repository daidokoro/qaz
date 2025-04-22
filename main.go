package main

import (
	"os"

	"github.com/daidokoro/qaz/cmd/root"
)

func main() {
	if err := root.Execute(); err != nil {
		os.Exit(-1)
	}
}
