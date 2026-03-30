package main

import (
	"fmt"
	"os"

	"github.com/andreasbaumgartner/gilo/internal/selfupdate"
	"github.com/andreasbaumgartner/gilo/internal/tui"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-update" || os.Args[1] == "--update") {
		if err := selfupdate.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
