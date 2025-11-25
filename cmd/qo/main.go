package main

import (
	"os"

	"github.com/kiki-ki/go-qo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
