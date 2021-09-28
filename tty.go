package tea

func (p *Program) initTerminal() error {
	err := p.initInput()
	if err != nil {
		return err
	}

	if p.console != nil {
		err = p.console.SetRaw()
		if err != nil {
			return err
		}
	}

	hideCursor(p.output)
	return nil
}

func (p Program) restoreTerminal() error {
	showCursor(p.output)

	if p.console != nil {
		err := p.console.Reset()
		if err != nil {
			return err
		}
	}

	return p.restoreInput()
}
