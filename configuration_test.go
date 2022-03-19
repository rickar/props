// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"errors"
	"io/fs"
	"math"
	"testing"
	"testing/fstest"
	"time"
)

var memFs = make(fstest.MapFS)

func init() {
	memFs["testapp.properties"] = &fstest.MapFile{
		Data: []byte(`
key1=abc
key2=def
key3=ghi
keyexp=${key2}
`),
	}
	memFs["testapp-test.properties"] = &fstest.MapFile{
		Data: []byte(`
key2=123
key3=
`),
	}
	memFs["testapp-prod.properties"] = &fstest.MapFile{
		Data: []byte(`
key2=456
`),
	}
}

func TestNewConfigurationNoProfile(t *testing.T) {
	c, err := NewConfiguration(memFs, "testapp")
	if err != nil {
		t.Errorf("got error: %v", err)
	}

	val, _ := c.Get("key1")
	if val != "abc" {
		t.Errorf("want: 'abc'; got: '%s'", val)
	}

	val, _ = c.Get("key2")
	if val != "def" {
		t.Errorf("want: 'def'; got: '%s'", val)
	}

	val, _ = c.Get("key3")
	if val != "ghi" {
		t.Errorf("want: 'ghi'; got '%s'", val)
	}

	val, _ = c.Get("key4")
	if val != "" {
		t.Errorf("want: ''; got: '%s'", val)
	}

	val, _ = c.Get("keyexp")
	if val != "def" {
		t.Errorf("want: 'def'; got: '%s'", val)
	}

	val = c.GetDefault("keyexp", "none")
	if val != "def" {
		t.Errorf("want: 'def'; got: '%s'", val)
	}

	val = c.GetDefault("keynone", "none")
	if val != "none" {
		t.Errorf("want: 'none'; got: '%s'", val)
	}
}

func TestNewConfigurationProfile(t *testing.T) {
	c, err := NewConfiguration(memFs, "testapp", "prod", "test", "bad")
	if err != nil {
		t.Errorf("got error: %v", err)
	}

	val, _ := c.Get("key1")
	if val != "abc" {
		t.Errorf("want: 'abc'; got: '%s'", val)
	}

	val, _ = c.Get("key2")
	if val != "456" {
		t.Errorf("want: '456'; got: '%s'", val)
	}

	val, _ = c.Get("key3")
	if val != "" {
		t.Errorf("want: ''; got '%s'", val)
	}

	val, _ = c.Get("key4")
	if val != "" {
		t.Errorf("want: ''; got: '%s'", val)
	}

	val, _ = c.Get("keyexp")
	if val != "456" {
		t.Errorf("want: '456'; got '%s'", val)
	}
}

type badFileInfo struct{}

func (b *badFileInfo) Name() string       { return "file" }
func (b *badFileInfo) Size() int64        { return 1024 }
func (b *badFileInfo) Mode() fs.FileMode  { return 0 }
func (b *badFileInfo) ModTime() time.Time { return time.Now() }
func (b *badFileInfo) IsDir() bool        { return false }
func (b *badFileInfo) Sys() interface{}   { return nil }

type badFile struct{}

func (b *badFile) Stat() (fs.FileInfo, error) { return &badFileInfo{}, nil }
func (b *badFile) Read([]byte) (int, error)   { return 0, errors.New("bad file read") }
func (b *badFile) Close() error               { return nil }

type badFs struct{}

func (b *badFs) Open(name string) (fs.File, error) {
	if name == "bad-read.properties" {
		return &badFile{}, nil
	}
	return nil, errors.New("bad file")
}

func (b *badFs) Stat(name string) (fs.FileInfo, error) {
	return &badFileInfo{}, nil
}

func TestNewConfigurationBadFile(t *testing.T) {
	_, err := NewConfiguration(&badFs{}, "bad", "open", "read")
	if err == nil {
		t.Errorf("want err; got none")
	}

	_, err = NewConfiguration(&badFs{}, "bad", "read")
	if err == nil {
		t.Errorf("want err; got none")
	}

	_, err = NewConfiguration(&badFs{}, "bad-open")
	if err == nil {
		t.Errorf("want err; got none")
	}

	_, err = NewConfiguration(&badFs{}, "bad-read")
	if err == nil {
		t.Errorf("want err; got none")
	}
}

func TestParseInt(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: NewExpander(p)}

	p.Set("bad", "abc")
	p.Set("badextra", "42x")
	p.Set("badfloat", "42.123")
	p.Set("good", "42")
	tests := []struct {
		key     string
		defVal  int
		want    int
		wantErr bool
	}{
		{"none", 123, 123, false},
		{"bad", 123, 123, true},
		{"badextra", 123, 123, true},
		{"badfloat", 123, 123, true},
		{"good", 123, 42, false},
	}

	for i, test := range tests {
		got, gotErr := c.ParseInt(test.key, test.defVal)
		if got != test.want || (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("[%d-%s] got: %d, gotErr: %v; want: %d, wantErr: %t", i, test.key, got, gotErr, test.want, test.wantErr)
		}
	}
}

func TestParseFloat(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: NewExpander(p)}

	p.Set("bad", "abc")
	p.Set("badextra", "42x")
	p.Set("good", "42.123")
	tests := []struct {
		key     string
		defVal  float64
		want    float64
		wantErr bool
	}{
		{"none", 123, 123, false},
		{"bad", 123, 123, true},
		{"badextra", 123, 123, true},
		{"good", 123, 42.123, false},
	}

	for i, test := range tests {
		got, gotErr := c.ParseFloat(test.key, test.defVal)
		if got != test.want || (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("[%d-%s] got: %f, gotErr: %v; want: %f, wantErr: %t", i, test.key, got, gotErr, test.want, test.wantErr)
		}
	}
}

func TestParseByteSize(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: NewExpander(p)}

	p.Set("bad1", "abc")
	p.Set("bad2", "42x")
	p.Set("bad3", "42_M")
	p.Set("bad4", "42M 50k")
	p.Set("bad5", "999999999999999999999999999999Ei")
	p.Set("bad6", "42...5Ti")
	p.Set("bad7", "42.5x")
	p.Set("goodint", "42")
	p.Set("goodfloat", "42.5")
	p.Set("goodk", "720k")
	p.Set("goodKi", "720 Ki")
	p.Set("goodm", "1.44M")
	p.Set("goodmi", "1.44 Mi")
	p.Set("goodg", "1.21G")
	p.Set("goodgi", "1.21 Gi")
	p.Set("goodt", "15T")
	p.Set("goodti", "15 Ti")
	p.Set("goodp", "20P")
	p.Set("goodpi", "20 Pi")
	p.Set("goode", "15E")
	p.Set("goodei", "15 Ei")

	tests := []struct {
		key     string
		defVal  uint64
		want    uint64
		wantErr bool
	}{
		{"none", 123, 123, false},
		{"bad1", 123, 123, true},
		{"bad2", 123, 123, true},
		{"bad3", 123, 123, true},
		{"bad4", 123, 123, true},
		{"bad5", 123, 123, true},
		{"bad6", 123, 123, true},
		{"bad7", 123, 123, true},
		{"goodint", 123, 42, false},
		{"goodfloat", 123, 43, false},
		{"goodk", 123, 720_000, false},
		{"goodKi", 123, 737_280, false},
		{"goodm", 123, 1_440_000, false},
		{"goodmi", 123, 1_509_949, false},
		{"goodg", 123, 1_210_000_000, false},
		{"goodgi", 123, 1_299_227_607, false},
		{"goodt", 123, 15_000_000_000_000, false},
		{"goodti", 123, 16_492_674_416_640, false},
		{"goodp", 123, 20_000_000_000_000_000, false},
		{"goodpi", 123, 22_517_998_136_852_480, false},
		{"goode", 123, 15_000_000_000_000_000_000, false},
		{"goodei", 123, 17_293_822_569_102_704_640, false},
	}

	for i, test := range tests {
		got, gotErr := c.ParseByteSize(test.key, test.defVal)
		if got != test.want || (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("[%d-%s] got: %d, gotErr: %v; want: %d, wantErr: %t", i, test.key, got, gotErr, test.want, test.wantErr)
		}
	}
}

func TestParseSize(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: NewExpander(p)}

	p.Set("bad1", "abc")
	p.Set("bad2", "42x")
	p.Set("bad3", "42_M")
	p.Set("bad4", "42M 50k")
	p.Set("bad5", "1e9k")
	p.Set("bad6", "42...5T")
	p.Set("bad7", "42.5x")
	p.Set("goodint", "42")
	p.Set("goodfloat", "42.5")
	p.Set("goodY", "1.23Y")
	p.Set("goodZ", "1.23 Z")
	p.Set("goodE", "1.23E")
	p.Set("goodP", "1.23 P")
	p.Set("goodT", "1.23T")
	p.Set("goodG", "1.23 G")
	p.Set("goodM", "1.23M")
	p.Set("goodk", "1.23 k")
	p.Set("goodh", "1.23h")
	p.Set("goodda", "1.23 da")
	p.Set("goodd", "1.23d")
	p.Set("goodc", "1.23 c")
	p.Set("goodm", "1.23m")
	p.Set("goodu", "1.23 u")
	p.Set("goodn", "1.23n")
	p.Set("goodp", "1.23 p")
	p.Set("goodf", "1.23f")
	p.Set("gooda", "1.23 a")
	p.Set("goodz", "1.23z")
	p.Set("goody", "1.23 y")

	tests := []struct {
		key     string
		defVal  float64
		want    float64
		wantErr bool
	}{
		{"none", 123, 123, false},
		{"bad1", 123, 123, true},
		{"bad2", 123, 123, true},
		{"bad3", 123, 123, true},
		{"bad4", 123, 123, true},
		{"bad5", 123, 123, true},
		{"bad6", 123, 123, true},
		{"bad7", 123, 123, true},
		{"goodint", 123, 42, false},
		{"goodfloat", 123, 42.5, false},
		{"goodY", 123, 1.23e24, false},
		{"goodZ", 123, 1.23e21, false},
		{"goodE", 123, 1.23e18, false},
		{"goodP", 123, 1.23e15, false},
		{"goodT", 123, 1.23e12, false},
		{"goodG", 123, 1.23e9, false},
		{"goodM", 123, 1.23e6, false},
		{"goodk", 123, 1.23e3, false},
		{"goodh", 123, 1.23e2, false},
		{"goodda", 123, 1.23e1, false},
		{"goodd", 123, 1.23e-1, false},
		{"goodc", 123, 1.23e-2, false},
		{"goodm", 123, 1.23e-3, false},
		{"goodu", 123, 1.23e-6, false},
		{"goodn", 123, 1.23e-9, false},
		{"goodp", 123, 1.23e-12, false},
		{"goodf", 123, 1.23e-15, false},
		{"gooda", 123, 1.23e-18, false},
		{"goodz", 123, 1.23e-21, false},
		{"goody", 123, 1.23e-24, false},
	}

	for i, test := range tests {
		got, gotErr := c.ParseSize(test.key, test.defVal)
		if math.Abs(got-test.want) > 0.0001 || (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("[%d-%s] got: %e, gotErr: %v; want: %e, wantErr: %t", i, test.key, got, gotErr, test.want, test.wantErr)
		}
	}
}

func TestParseBool(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: NewExpander(p)}

	p.Set("bad", "abc")
	p.Set("badextra", "truex")
	p.Set("goodtrue", "true")
	p.Set("goodfalse", "false")

	p.Set("bad", "def")
	p.Set("badextra", "12")
	p.Set("good1", "true")
	p.Set("good2", "t")
	p.Set("good3", "yes")
	p.Set("good4", "y")
	p.Set("good5", "1")
	p.Set("good6", "on")
	p.Set("good7", "false")
	p.Set("good8", "f")
	p.Set("good9", "no")
	p.Set("good10", "n")
	p.Set("good11", "0")
	p.Set("good12", "off")

	tests := []struct {
		key     string
		defVal  bool
		want    bool
		wantErr bool
	}{
		{"none", true, true, false},
		{"bad", true, true, true},
		{"badextra", true, true, true},
		{"goodtrue", false, true, false},
		{"goodfalse", true, false, false},

		{"bad", true, true, true},
		{"badextra", true, true, true},
		{"good1", false, true, false},
		{"good2", false, true, false},
		{"good3", false, true, false},
		{"good4", false, true, false},
		{"good5", false, true, false},
		{"good6", false, true, false},
		{"good7", true, false, false},
		{"good8", true, false, false},
		{"good9", true, false, false},
		{"good10", true, false, false},
		{"good11", true, false, false},
		{"good12", true, false, false},
	}

	for i, test := range tests {
		c.StrictBool = i < 5
		got, gotErr := c.ParseBool(test.key, test.defVal)
		if got != test.want || (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("[%d-%s] got: %t, gotErr: %v; want: %t, wantErr: %t", i, test.key, got, gotErr, test.want, test.wantErr)
		}
	}
}

func TestParseDuration(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: NewExpander(p)}

	p.Set("bad", "abc")
	p.Set("badextra", "42x")
	p.Set("badfloat", "42.123")
	p.Set("good", "42h")
	tests := []struct {
		key     string
		defVal  time.Duration
		want    time.Duration
		wantErr bool
	}{
		{"none", time.Duration(123), time.Duration(123), false},
		{"bad", time.Duration(123), time.Duration(123), true},
		{"badextra", time.Duration(123), time.Duration(123), true},
		{"badfloat", time.Duration(123), time.Duration(123), true},
		{"good", time.Duration(123), 42 * time.Hour, false},
	}

	for i, test := range tests {
		got, gotErr := c.ParseDuration(test.key, test.defVal)
		if got != test.want || (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("[%d-%s] got: %d, gotErr: %v; want: %d, wantErr: %t", i, test.key, got, gotErr, test.want, test.wantErr)
		}
	}
}

func TestParseDate(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: NewExpander(p)}

	p.Set("bad", "abc")
	p.Set("badextra", "2000-01-01x")
	p.Set("badformat", "99-12-31")
	p.Set("good", "2000-01-01")
	p.Set("goodformat", "2001.05.04 12:30")
	tests := []struct {
		key     string
		defVal  time.Time
		want    time.Time
		wantErr bool
	}{
		{"none", time.Time{}, time.Time{}, false},
		{"bad", time.Time{}, time.Time{}, true},
		{"badextra", time.Time{}, time.Time{}, true},
		{"badformat", time.Time{}, time.Time{}, true},
		{"good", time.Time{}, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{"goodformat", time.Time{}, time.Date(2001, 5, 4, 12, 30, 0, 0, time.UTC), false},
	}

	for i, test := range tests {
		if test.key == "goodformat" {
			c.DateFormat = "2006.01.02 15:04"
		} else {
			c.DateFormat = ""
		}
		got, gotErr := c.ParseDate(test.key, test.defVal)
		if !got.Equal(test.want) || (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("[%d-%s] got: %v, gotErr: %v; want: %v, wantErr: %t", i, test.key, got, gotErr, test.want, test.wantErr)
		}
	}
}

func TestConfigDecrypt(t *testing.T) {
	p := NewProperties()
	c := &Configuration{Props: p}

	p.Set("novalue", "[enc:0]")
	p.Set("notencrypted", "[enc:0]plaintext")
	p.Set("aesgcm.none", "[enc:1]")
	p.Set("aesgcm.badbase64", "[enc:1]$$$$")
	p.Set("aesgcm.good", "[enc:1]qzp-F5Cw1G5n9c7D5LFw4NzanvZzdQzONRPlM3JpLLCO6swpBQ==")
	p.Set("aesgcm.good256", "[enc:1]x-f72AdaTiZfRvX_6kekHpd4xUj1JmMFwbwiADbMZXgJjGpJNg==")
	p.Set("badalg", "[enc:x]abcd")
	p.Set("noalg", "abcd")

	badpass := "123"
	pass := "1234567890123456"
	pass256 := "12345678901234567890123456789012"

	val, err := c.Decrypt(badpass, "none", "default")
	if val != "default" || err != nil {
		t.Errorf("want: 'default', err == nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(badpass, "novalue", "default")
	if val != "" || err != nil {
		t.Errorf("want: '', err == nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(badpass, "notencrypted", "default")
	if val != "plaintext" || err != nil {
		t.Errorf("want: 'plaintext', err == nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(badpass, "aesgcm.good", "default")
	if val != "default" || err == nil {
		t.Errorf("want: 'default', err != nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(pass, "aesgcm.none", "default")
	if val != "default" || err == nil {
		t.Errorf("want: 'default', err != nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(pass, "aesgcm.badbase64", "default")
	if val != "default" || err == nil {
		t.Errorf("want: 'default', err != nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(pass, "aesgcm.good", "default")
	if val != "plaintext" || err != nil {
		t.Errorf("want: 'plaintext', err == nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(pass, "aesgcm.good256", "default")
	if val != "default" || err == nil {
		t.Errorf("want: 'default', err != nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(pass256, "aesgcm.good256", "default")
	if val != "plaintext" || err != nil {
		t.Errorf("want: 'plaintext', err == nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(pass, "badalg", "default")
	if val != "default" || err == nil {
		t.Errorf("want: 'default', err != nil; got: %s, %v", val, err)
	}

	val, err = c.Decrypt(pass, "noalg", "default")
	if val != "default" || err == nil {
		t.Errorf("want: 'default', err != nil; got: %s, %v", val, err)
	}
}
