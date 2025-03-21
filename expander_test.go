// (c) 2013 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"reflect"
	"sort"
	"testing"
)

func TestNewExpander(t *testing.T) {
	l := &Properties{}
	e := NewExpander(l)
	if len(e.Prefix) < 1 {
		t.Error("want prefix; got none")
	}
	if len(e.Suffix) < 1 {
		t.Error("want suffix; got none")
	}
	if e.Source != l {
		t.Error("want Source=l; got mismatch")
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
	p := NewProperties()
	for _, test := range noExpand {
		p.Set(test.key, test.val)
	}
	e := NewExpander(p)

	for _, test := range noExpand {
		got, ok := e.Get(test.key)
		if got != test.want || !ok {
			t.Errorf("want: %q; got: %q, %t", test.want, got, ok)
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
	p := NewProperties()
	p.Set("one", "1")
	p.Set("two", "2")

	for _, test := range singleExpand {
		p.Set(test.key, test.val)
	}

	e := NewExpander(p)

	for _, test := range singleExpand {
		got, ok := e.Get(test.key)
		if got != test.want || !ok {
			t.Errorf("want: %q; got: %q, %t", test.want, got, ok)
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
	{"key7", "foo${cycle}bar", "foo${cycle}bar"},
}

func TestNestExpand(t *testing.T) {
	p := NewProperties()

	p.Set("one", "1")
	p.Set("two", "2")
	p.Set("four", "4")
	p.Set("one2", "A")
	p.Set("three4", "B")
	p.Set("twoB", "C")
	p.Set("oneC", "D")
	p.Set("exp", "${exp2}")
	p.Set("exp2", "${exp3}")
	p.Set("exp3", "ZZZ")
	p.Set("recurse", "${recurse}")
	p.Set("cycle", "${cycle2}")
	p.Set("cycle2", "${cycle3}")
	p.Set("cycle3", "${cycle}")

	for _, test := range nestExpand {
		p.Set(test.key, test.val)
	}

	e := NewExpander(p)

	for _, test := range nestExpand {
		got, ok := e.Get(test.key)
		if got != test.want || !ok {
			t.Errorf("want: %q; got: %q; %t", test.want, got, ok)
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
	p := NewProperties()
	p.Set("one", "1")

	for _, test := range sameToken {
		p.Set(test.key, test.val)
	}

	e := NewExpander(p)
	e.Prefix = "@"
	e.Suffix = "@"

	for _, test := range sameToken {
		got, ok := e.Get(test.key)
		if got != test.want || !ok {
			t.Errorf("want: %q; got: %q, %t", test.want, got, ok)
		}
	}
}

func TestDefault(t *testing.T) {
	p := NewProperties()
	p.Set("key", "val")

	e := NewExpander(p)

	if v := e.GetDefault("key", "none"); v != "val" {
		t.Errorf("want: val; got: %s", v)
	}
	if v := e.GetDefault("key2", "none"); v != "none" {
		t.Errorf("want: none; got: %s", v)
	}
}

var limits = []expTest{
	{"key1", "foo${one}bar", "foo1bar"},
	{"key2", "foo${two}bar", "foo20bar"},
	{"key3", "foo${three}bar", "foo${thirtyOne}bar"},
}

func TestLimit(t *testing.T) {
	p := NewProperties()
	p.Set("one", "1")

	p.Set("two", "${twenty}")
	p.Set("twenty", "20")

	p.Set("three", "${thirty}")
	p.Set("thirty", "${thirtyOne}")
	p.Set("thirtyOne", "31")

	for _, test := range limits {
		p.Set(test.key, test.val)
	}

	e := NewExpander(p)
	e.Limit = 2

	for _, test := range limits {
		got, ok := e.Get(test.key)
		if got != test.want || !ok {
			t.Errorf("want: %q; got: %q, %t", test.want, got, ok)
		}
	}
}

func TestExpanderNames(t *testing.T) {
	p := NewProperties()
	p.Set("one", "1")

	p.Set("two", "${twenty}")
	p.Set("twenty", "20")

	p.Set("three", "${thirty}")
	p.Set("thirty", "${thirtyOne}")
	p.Set("thirtyOne", "31")

	e := NewExpander(p)

	got := e.Names()
	sort.Strings(got)
	want := []string{"one", "three", "thirty", "thirtyOne", "two", "twenty"}
	if reflect.DeepEqual(got, want) {
		t.Errorf("want: %v; got: %v", want, got)
	}
}
