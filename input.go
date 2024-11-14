package tea

import (
	"fmt"
	"strings"
)

// UnknownMsg represents an unknown message.
type UnknownMsg string

// String returns a string representation of the unknown message.
func (e UnknownMsg) String() string {
	return fmt.Sprintf("%q", string(e))
}

// multiMsg represents multiple messages event.
type multiMsg []Msg

// String returns a string representation of the multiple messages event.
func (e multiMsg) String() string {
	var sb strings.Builder
	for _, ev := range e {
		sb.WriteString(fmt.Sprintf("%v\n", ev))
	}
	return sb.String()
}
