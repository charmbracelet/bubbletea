package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/tea"
	"github.com/charmbracelet/teaparty/pager"
)

func main() {
	content, err := ioutil.ReadFile("artichoke.md")
	if err != nil {
		fmt.Println("could not load file:", err)
		os.Exit(1)
	}

	tea.AltScreen()
	defer tea.ExitAltScreen()
	if err := pager.NewProgram(string(content)).Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
