// (c) 2013 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"bytes"
	"io"
	"reflect"
	"sort"
	"testing"
)

func TestNewProps(t *testing.T) {
	p := NewProperties()
	if len(p.values) > 0 {
		t.Errorf("want: 0 elements; got: %d", len(p.values))
	}
}

var comments = `
# line 1
! line 2
   # line 3
   ! line 4
  # line 5
  ! line 6
`

func TestReadComments(t *testing.T) {
	p, err := Read(bytes.NewBufferString(comments))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	if len(p.values) > 0 {
		t.Errorf("want: 0 elements; got: %d", len(p.values))
	}
}

var simple = `
key1=a
key2=b
key3=c
`

func TestReadSimple(t *testing.T) {
	p, err := Read(bytes.NewBufferString(simple))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key1": "a",
		"key2": "b",
		"key3": "c",
	}

	if !reflect.DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}
}

var continued = `
key1=abc\
    	def
key\
	2\
	3 = ghi\
	j\
	k\
	l
`

func TestReadContinued(t *testing.T) {
	p, err := Read(bytes.NewBufferString(continued))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key1":  "abcdef",
		"key23": "ghijkl",
	}

	if !reflect.DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}
}

var keys = `
key1=a
key2:b
key3 c
key4 = d
key5 : e
key6   f
key7
key8=g
key9=
key10
key11`

func TestReadKeys(t *testing.T) {
	p, err := Read(bytes.NewBufferString(keys))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key1":  "a",
		"key2":  "b",
		"key3":  "c",
		"key4":  "d",
		"key5":  "e",
		"key6":  "f",
		"key7":  "",
		"key8":  "g",
		"key9":  "",
		"key10": "",
		"key11": "",
	}

	if !reflect.DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}

	if _, ok := p.values["key7"]; !ok {
		t.Error("want: key7; got none")
	}
}

var escapes = `\key0=123
key\n1=a\nb\n
key\t2:c\td
key\f3 e\ff
key\\4=g\\h
key\r5:i\rj
key\z6 k\3l
key\u005a7=m\u2126n
key\uuu00478=o\uzp
key\uD834\uDD1E9=q\uD800\uDC00r
key\
    \f10=s\
	\ft
key11=\u
key12=\uZ
key13 \t =abc
key14     
`

func TestReadEscapes(t *testing.T) {
	p, err := Read(bytes.NewBufferString(escapes))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key0":    "123",
		"key\n1":  "a\nb\n",
		"key\t2":  "c\td",
		"key\f3":  "e\ff",
		"key\\4":  "g\\h",
		"key\r5":  "i\rj",
		"keyz6":   "k3l",
		"keyZ7":   "mΩn",
		"keyG8":   "o\uFFFDp",
		"key𝄞9":   "q𐀀r",
		"key\f10": "s\ft",
		"key11":   "\uFFFD",
		"key12":   "\uFFFD",
		"key13":   "\t =abc",
		"key14":   "",
	}

	if !reflect.DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}
}

type ErrorReader struct {
}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrNoProgress
}

func TestReadError(t *testing.T) {
	p, err := Read(&ErrorReader{})
	if p != nil || err == nil {
		t.Errorf("want err; got none")
	}
}

func TestGet(t *testing.T) {
	p := NewProperties()
	p.values["key1"] = "foo"

	if v, ok := p.Get("key1"); v != "foo" || !ok {
		t.Errorf("want: foo; got: %q, %t", v, ok)
	}

	if v, ok := p.Get("key2"); v != "" || ok {
		t.Errorf("want: \"\"; got: %q, %t", v, ok)
	}
}

func TestGetDefault(t *testing.T) {
	p := NewProperties()
	p.values["key1"] = "foo"

	if p.GetDefault("key1", "bar") != "foo" {
		t.Errorf("want: foo; got: %q", p.GetDefault("key1", "bar"))
	}

	if p.GetDefault("key2", "bar") != "bar" {
		t.Errorf("want: bar; got: %q", p.GetDefault("key2", "bar"))
	}
}

func TestSet(t *testing.T) {
	p := NewProperties()
	p.Set("key1", "foo")
	p.Set("key1", "bar")

	if p.values["key1"] != "bar" {
		t.Errorf("want: bar; got %q", p.values["key1"])
	}
}

func TestClear(t *testing.T) {
	p := NewProperties()
	p.Set("key1", "foo")
	p.Set("key2", "bar")

	if len(p.values) != 2 {
		t.Errorf("want len: 2, got: %d", len(p.values))
	}

	p.Clear()
	if len(p.values) != 0 {
		t.Errorf("want len: 0, got: %d", len(p.values))
	}
}

func TestNames(t *testing.T) {
	p := NewProperties()
	p.values["key1"] = "foo"
	p.values["key2"] = "bar"

	want := []string{"key1", "key2"}
	got := p.Names()

	sort.Strings(want)
	sort.Strings(got)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %#v, got: %#v", want, got)
	}
}

var writeTests = []struct {
	key  string
	val  string
	want string
}{
	{"key", "val", "key=val\n"},
	{"key", "  foo bar baz", "key=\\ \\ foo bar baz\n"},
	{"key:=#!", ":=#!foo bar baz", "key\\:\\=\\#\\!=\\:\\=\\#\\!foo bar baz\n"},
	{"key foo", "bar", "key\\ foo=bar\n"},
	{"key\nfoo", "bar\nbaz", "key\\nfoo=bar\\nbaz\n"},
	{"key\rfoo", "bar\rbaz", "key\\rfoo=bar\\rbaz\n"},
	{"key\ffoo", "bar\fbaz", "key\\ffoo=bar\\fbaz\n"},
	{"key\tfoo", "bar\tbaz", "key\\tfoo=bar\\tbaz\n"},
	{"key\u00A0foo", "bar\u00A9baz", "key\\u00a0foo=bar\\u00a9baz\n"},
}

func TestWrite(t *testing.T) {
	for _, test := range writeTests {
		p := NewProperties()
		p.values[test.key] = test.val

		buf := new(bytes.Buffer)
		err := p.Write(buf)
		if err != nil {
			t.Errorf("got err: %v", err)
		}

		got := buf.String()
		if got != test.want {
			t.Errorf("want: %q; got: %q", test.want, got)
		}
	}
}

type ErrorWriter struct {
}

func (e *ErrorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}

func TestWriteError(t *testing.T) {
	p := NewProperties()
	p.Set("key1", "foo")
	p.Set("key2", "bar")

	err := p.Write(&ErrorWriter{})
	if err == nil {
		t.Errorf("want err; got none")
	}
}
