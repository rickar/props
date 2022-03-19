// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"bytes"
	"testing"
)

func TestFinishUtfEscape_BadHexChar(t *testing.T) {
	s := &scanner{
		p: &Properties{},
	}
	s.current = &s.value
	s.utfUnits = make([]bytes.Buffer, 1)
	s.utfUnits[0].WriteRune('z')
	s.finishUtfEscape()

	r, _, _ := s.current.ReadRune()
	if r != '\uFFFD' {
		t.Errorf("want: \uFFFD; got: %x", r)
	}
}
