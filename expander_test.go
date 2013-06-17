// (c) 2013 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"testing"
)

func TestNewExpander(t *testing.T) {
	e := NewExpander()
	if len(e.values) > 0 {
		t.Errorf("want: 0 elements; got: %d", len(e.values))
	}
	if len(e.Prefix) < 1 {
		t.Error("want prefix; got none")
	}
	if len(e.Suffix) < 1 {
		t.Error("want suffix; got none")
	}
}

type expTest struct {
	key  string
	val  string
	want string
}

var noExpand = []expTest{
	{"key1", "foo", "foo"},
	{"key2", "bar${", "bar${"},
	{"key3", "baz}", "baz}"},
}

func TestNoExpand(t *testing.T) {
	e := NewExpander()

	for _, test := range noExpand {
		e.Set(test.key, test.val)
	}

	for _, test := range noExpand {
		got := e.Get(test.key)
		if got != test.want {
			t.Errorf("want: %q; got: %q", test.want, got)
		}
	}
}

var singleExpand = []expTest{
	{"key1", "foo${one}bar", "foo1bar"},
	{"key2", "${one}foobar", "1foobar"},
	{"key3", "foobar${one}", "foobar1"},
	{"key4", "${one}", "1"},
	{"key5", "foo${one}${two}bar", "foo12bar"},
	{"key6", "${one}${two}foobar", "12foobar"},
	{"key7", "foobar${one}${two}", "foobar12"},
	{"key8", "foo${one}bar${two}", "foo1bar2"},
	{"key9", "foo${zzz}bar", "foo${zzz}bar"},
	{"key10", "${zzz}foobar", "${zzz}foobar"},
	{"key11", "foobar${zzz}", "foobar${zzz}"},
	{"key12", "${zzz}", "${zzz}"},
}

func TestSingleExpand(t *testing.T) {
	e := NewExpander()

	e.Set("one", "1")
	e.Set("two", "2")

	for _, test := range singleExpand {
		e.Set(test.key, test.val)
	}

	for _, test := range singleExpand {
		got := e.Get(test.key)
		if got != test.want {
			t.Errorf("want: %q; got: %q", test.want, got)
		}
	}
}

var nestExpand = []expTest{
	{"key1", "foo${one${two}}bar", "fooAbar"},
	{"key2", "${one${two}}foobar", "Afoobar"},
	{"key3", "foobar${one${two}}", "foobarA"},
	{"key4", "foobar${one${two${three${four}}}}", "foobarD"},
	{"key5", "foo${exp}bar", "fooZZZbar"},
	{"key6", "foo${recurse}bar", "foo${recurse}bar"},
}

func TestNestExpand(t *testing.T) {
	e := NewExpander()

	e.Set("one", "1")
	e.Set("two", "2")
	e.Set("four", "4")
	e.Set("one2", "A")
	e.Set("three4", "B")
	e.Set("twoB", "C")
	e.Set("oneC", "D")
	e.Set("exp", "${exp2}")
	e.Set("exp2", "${exp3}")
	e.Set("exp3", "ZZZ")
	e.Set("recurse", "${recurse}")

	for _, test := range nestExpand {
		e.Set(test.key, test.val)
	}

	for _, test := range nestExpand {
		got := e.Get(test.key)
		if got != test.want {
			t.Errorf("want: %q; got: %q", test.want, got)
		}
	}
}

var sameToken = []expTest{
	{"key1", "foo@one@bar", "foo1bar"},
	{"key2", "@one@foobar", "1foobar"},
	{"key3", "foobar@one@", "foobar1"},
	{"key4", "foo@one@two@@", "foo1two@@"},
}

func TestSameToken(t *testing.T) {
	e := NewExpander()
	e.Prefix = "@"
	e.Suffix = "@"

	e.Set("one", "1")

	for _, test := range sameToken {
		e.Set(test.key, test.val)
	}

	for _, test := range sameToken {
		got := e.Get(test.key)
		if got != test.want {
			t.Errorf("want: %q; got: %q", test.want, got)
		}
	}
}
