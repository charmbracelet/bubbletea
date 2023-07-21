package tea

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stringMsg string

func TestEvery(t *testing.T) {
	expected := stringMsg("every ms")
	msg := Every(time.Millisecond, func(t time.Time) Msg {
		return expected
	})()
	assert.Equal(t, expected, msg)
}

func TestTick(t *testing.T) {
	expected := stringMsg("tick")
	msg := Tick(time.Millisecond, func(t time.Time) Msg {
		return expected
	})()
	assert.Equal(t, expected, msg)
}

type errorMsg struct {
	error
}

func TestSequentially(t *testing.T) {
	expectedErrMsg := errorMsg{error: errors.New("some err")}
	expectedStrMsg := stringMsg("some msg")

	nilReturnCmd := func() Msg {
		return nil
	}

	for name, test := range map[string]struct {
		cmds     []Cmd
		expected Msg
	}{
		"all nil": {
			cmds:     []Cmd{nilReturnCmd, nilReturnCmd},
			expected: nil,
		},
		"null cmds": {
			cmds:     []Cmd{nil, nil},
			expected: nil,
		},
		"one error": {
			cmds: []Cmd{
				nilReturnCmd,
				func() Msg {
					return expectedErrMsg
				},
				nilReturnCmd,
			},
			expected: expectedErrMsg,
		},
		"some msg": {
			cmds: []Cmd{
				nilReturnCmd,
				func() Msg {
					return expectedStrMsg
				},
				nilReturnCmd,
			},
			expected: expectedStrMsg,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Sequentially(test.cmds...)())
		})
	}
}

func TestBatch(t *testing.T) {
	for name, test := range map[string]struct {
		cmds        []Cmd
		expectedLen int
	}{
		"nil cmd": {
			cmds:        []Cmd{nil},
			expectedLen: 0,
		},
		"empty cmd": {
			cmds:        nil,
			expectedLen: 0,
		},
		"single cmd": {
			cmds:        []Cmd{Quit},
			expectedLen: 1,
		},
		"mixed nil cmds": {
			cmds:        []Cmd{nil, Quit, nil, Quit, nil, nil},
			expectedLen: 2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if test.expectedLen == 0 {
				assert.Nil(t, Batch(test.cmds...))
			} else {
				assert.Len(t, Batch(test.cmds...)(), test.expectedLen)
			}
		})
	}
}
