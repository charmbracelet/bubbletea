package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/boba"
	"github.com/charmbracelet/boba/pager"
)

func main() {
	content, err := ioutil.ReadFile("artichoke.md")
	if err != nil {
		fmt.Println("could not load file:", err)
		os.Exit(1)
	}

	boba.AltScreen()
	defer boba.ExitAltScreen()
	if err := pager.NewProgram(string(content)).Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
