package main

import (
	"testing"

	"github.com/knz/catwalk"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	catwalk.RunModel(t, "main_testdata", m)
}
