// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"testing"
)

func TestCombined(t *testing.T) {
	c := &Combined{}
	val, ok := c.Get("non.init")
	if ok || val != "" {
		t.Errorf("want: false, ''; got: %t, '%s'", ok, val)
	}

	p1 := NewProperties()
	p1.Set("props.test.val1", "abc")
	p1.Set("props.test.val2", "")
	p2 := NewProperties()
	p2.Set("props.test.val1", "def")
	p2.Set("props.test.val2", "")
	p2.Set("props.test.val3", "ghi")
	c.Sources = []PropertyGetter{p1, p2}

	val, ok = c.Get("props.test.val1")
	if !ok || val != "abc" {
		t.Errorf("want: true, 'abc'; got: %t, '%s'", ok, val)
	}

	val, ok = c.Get("props.test.val2")
	if !ok || val != "" {
		t.Errorf("want: true, ''; got: %t, '%s'", ok, val)
	}

	val, ok = c.Get("props.test.val3")
	if !ok || val != "ghi" {
		t.Errorf("want: true, 'ghi'; got %t, '%s'", ok, val)
	}

	val, ok = c.Get("props.test.val4")
	if ok || val != "" {
		t.Errorf("want: false, ''; got %t, '%s'", ok, val)
	}

	val = c.GetDefault("props.test.val1", "other")
	if val != "abc" {
		t.Errorf("want: 'abc'; got '%s'", val)
	}

	val = c.GetDefault("props.test.val4", "other")
	if val != "other" {
		t.Errorf("want: 'other'; got: '%s'", val)
	}
}
