// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"os"
	"testing"
)

func TestEnvironmentGet(t *testing.T) {
	e := &Environment{}

	os.Setenv("PROPS_TEST_VAL1", "abc")
	os.Setenv("PROPS_TEST_VAL2", "")

	val, ok := e.Get("PROPS_TEST_VAL1")
	if !ok || val != "abc" {
		t.Errorf("want: true, 'abc'; got: %t, '%s'", ok, val)
	}

	val, ok = e.Get("PROPS_TEST_VAL2")
	if !ok || val != "" {
		t.Errorf("want: true, ''; got: %t, '%s'", ok, val)
	}

	val, ok = e.Get("PROPS_TEST_VAL3")
	if ok || val != "" {
		t.Errorf("want: false, ''; got %t, '%s'", ok, val)
	}

	e.Normalize = true
	val, ok = e.Get("props.TEST:val1")
	if !ok || val != "abc" {
		t.Errorf("want: true, 'abc'; got: %t, '%s'", ok, val)
	}
}

func TestEnvironmentGetDefault(t *testing.T) {
	e := &Environment{}

	os.Setenv("PROPS_TEST_VAL1", "abc")
	os.Setenv("PROPS_TEST_VAL2", "")

	val := e.GetDefault("PROPS_TEST_VAL1", "zzz")
	if val != "abc" {
		t.Errorf("want: 'abc'; got: '%s'", val)
	}

	val = e.GetDefault("PROPS_TEST_VAL2", "def")
	if val != "" {
		t.Errorf("want: ''; got: '%s'", val)
	}

	val = e.GetDefault("PROPS_TEST_VAL3", "ghi")
	if val != "ghi" {
		t.Errorf("want: 'ghi'; got: '%s'", val)
	}
}

func TestEnvironmentNames(t *testing.T) {
	e := &Environment{}

	os.Setenv("PROPS_TEST_VAL1", "abc")
	os.Setenv("PROPS_TEST_VAL2", "")

	names := e.Names()
	var found1, found2 bool
	for _, v := range names {
		if v == "PROPS_TEST_VAL1" {
			found1 = true
		} else if v == "PROPS_TEST_VAL2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("name not found; found1: %t, found2: %t", found1, found2)
	}
}
