package main

/*
func TestApp(t *testing.T) {
	// TODO: Enable this test again
	// Since we added colorprofile.Writer to standard_renderer.go, this test
	// keeps failing. This is because the output is colored and has escape
	// sequences but the test runs against a buffer output and not a terminal,
	// tty, or pty. One way to fix this is to pass a color profile to the test
	// program using [tea.WithColorProfile(Ascii)].
	t.Skip("this test is currently disabled")

	m := model(10)
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
	)
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Fatal(err)
		}
	})

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Type("I'm typing things, but it'll be ignored by my program")
	tm.Send("ignored msg")
	tm.Send(tea.KeyPressMsg{
		Code: tea.KeyEnter,
	})

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}

	out := readBts(t, tm.FinalOutput(t))
	if !regexp.MustCompile(`This program will exit in \d+ seconds`).Match(out) {
		t.Fatalf("output does not match the given regular expression: %s", string(out))
	}
	teatest.RequireEqualOutput(t, out)

	if tm.FinalModel(t).(model) != 9 {
		t.Errorf("expected model to be 10, was %d", m)
	}
}

func TestAppInteractive(t *testing.T) {
	t.Skip("This test is flaky and needs to be fixed.\n" +
		"We need a more concrete way to set the initial terminal size")

	m := model(10)
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
	)

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Send("ignored msg")

	if bts := readBts(t, tm.Output()); !bytes.Contains(bts, []byte("This program will exit in 9 seconds")) {
		t.Fatalf("output does not match: expected %q", string(bts))
	}

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("This program will exit in 7 seconds"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyPressMsg{
		Code: tea.KeyEnter,
	})

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}

	if tm.FinalModel(t).(model) != 7 {
		t.Errorf("expected model to be 7, was %d", m)
	}
}

func readBts(tb testing.TB, r io.Reader) []byte {
	tb.Helper()
	bts, err := io.ReadAll(r)
	if err != nil {
		tb.Fatal(err)
	}
	return bts
}
*/
