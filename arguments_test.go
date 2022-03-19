// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"os"
	"testing"
)

func TestArgumentsGet(t *testing.T) {
	a := &Arguments{}
	os.Args = make([]string, 0)
	os.Args = append(os.Args, "prog")
	os.Args = append(os.Args, "--props.test.val1=abc")
	os.Args = append(os.Args, "--props.test.val2=")

	val, ok := a.Get("props.test.val1")
	if !ok || val != "abc" {
		t.Errorf("want: true, 'abc'; got: %t, '%s'", ok, val)
	}

	val, ok = a.Get("props.test.val2")
	if !ok || val != "" {
		t.Errorf("want: true, ''; got: %t, '%s'", ok, val)
	}

	val, ok = a.Get("props.test.val3")
	if ok || val != "" {
		t.Errorf("want: false, ''; got %t, '%s'", ok, val)
	}

	a.Prefix = "--props."
	val, ok = a.Get("test.val1")
	if !ok || val != "abc" {
		t.Errorf("want: true, 'abc'; got: %t, '%s'", ok, val)
	}
}

func TestArgumentsGetDefault(t *testing.T) {
	a := &Arguments{Prefix: "--props."}
	os.Args = make([]string, 0)
	os.Args = append(os.Args, "prog")
	os.Args = append(os.Args, "--props.test.val1=abc")
	os.Args = append(os.Args, "--props.test.val2=")

	val := a.GetDefault("test.val1", "zzz")
	if val != "abc" {
		t.Errorf("want: 'abc'; got: '%s'", val)
	}

	val = a.GetDefault("test.val2", "def")
	if val != "" {
		t.Errorf("want: ''; got: '%s'", val)
	}

	val = a.GetDefault("test.val3", "ghi")
	if val != "ghi" {
		t.Errorf("want: 'ghi'; got: '%s'", val)
	}
}
