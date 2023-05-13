//go:build js
// +build js

package tea

import (
	"errors"
	"os"
)

func (p *Program) initInput() error {
	return errors.New("unavailable in js")
}

func (p *Program) restoreInput() error {
	return errors.New("unavailable in js")
}

func openInputTTY() (*os.File, error) {
	return nil, errors.New("unavailable in js")
}
