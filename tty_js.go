//go:build js
// +build js

package tea

import (
	"errors"
	"os"
)

func (p *Program) initInput() error {
	return nil
}

func (p *Program) restoreInput() error {
	return nil
}

func openInputTTY() (*os.File, error) {
	return nil, errors.New("unavailable in js")
}
