// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"fmt"
	"io/fs"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// EncryptNone represents a value that has not yet been encrypted
	EncryptNone = "[enc:0]"
	// EncryptAESGCM represents a value that has been encryped with AES-GCM
	EncryptAESGCM = "[enc:1]"

	// EncryptDefault represents the default encryption algorithm
	EncryptDefault = EncryptAESGCM
)

// sizePattern provides the expression used for matching size values
var sizePattern = regexp.MustCompile("^([0-9.]+)\\s{0,1}([a-zA-Z]*)$")

// Configuration represents an application's configuration parameters provided
// by properties.
//
// It can be created directly or through NewConfiguration which provides
// configuration by convention.
type Configuration struct {
	// Props provides the configuration values to retrieve and/or parse.
	Props PropertyGetter

	// DateFormat provides the format string to use when parsing dates as
	// defined in the time package. If blank, the default 2006-01-02 is used.
	DateFormat string
	// StrictBool determines whether bool parsing is strict or not. When true,
	// only "true" and "false" values are considered valid. When false,
	// additional "boolean-like" values are accepted such as 0 and 1. See
	// ParseBool for details.
	StrictBool bool
}

// NewConfiguration creates a Configuration using common conventions.
//
// The returned Configuration uses an Expander to return properties in the
// following priority order:
//   1. Command line arguments
//   2. Environment variables
//   3. <prefix>-<profile>.properties for the provided prefix and profiles
//      values (in order)
//   4. <prefix>.properties for the provided prefix value
// The first matching property value found will be returned.
//
// An error will be returned if one of the property files could not be read or
// parsed.
func NewConfiguration(fileSys fs.StatFS, prefix string, profiles ...string) (*Configuration, error) {
	c := &Combined{}
	c.Sources = make([]PropertyGetter, 0)
	c.Sources = append(c.Sources, &Arguments{})
	c.Sources = append(c.Sources, &Environment{Normalize: true})

	for _, profile := range profiles {
		filename := prefix + "-" + profile + ".properties"
		stat, err := fileSys.Stat(filename)
		if err == nil && !stat.IsDir() {
			p := NewProperties()
			f, err := fileSys.Open(filename)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			err = p.Load(f)
			if err != nil {
				return nil, err
			}
			c.Sources = append(c.Sources, p)
		}
	}

	filename := prefix + ".properties"
	stat, err := fileSys.Stat(filename)
	if err == nil && !stat.IsDir() {
		p := NewProperties()
		f, err := fileSys.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		err = p.Load(f)
		if err != nil {
			return nil, err
		}
		c.Sources = append(c.Sources, p)
	}
	return &Configuration{Props: NewExpander(c)}, nil
}

// Get retrieves the value of a property. If the property does not exist, an
// empty string will be returned. The bool return value indicates whether
// the property was found.
func (c *Configuration) Get(key string) (string, bool) {
	return c.Props.Get(key)
}

// GetDefault retrieves the value of a property. If the property does not
// exist, then the default value will be returned.
func (c *Configuration) GetDefault(key, defVal string) string {
	return c.Props.GetDefault(key, defVal)
}

// ParseInt converts a property value to an int. If the property does not exist,
// then the default value will be returned with a nil error. If the property
// value could not be parsed, then an error and the default value will be
// returned.
func (c *Configuration) ParseInt(key string, defVal int) (int, error) {
	val, ok := c.Props.Get(key)
	if ok {
		var err error
		result, err := strconv.Atoi(val)
		if err != nil {
			return defVal, fmt.Errorf("invalid int value %s=%s [%w]", key, val, err)
		}
		return result, nil
	} else {
		return defVal, nil
	}
}

// ParseFloat converts a property value to a float64. If the property does not
// exist, then the default value will be returned with a nil error. If the
// property value could not be parsed, then an error and the default value will
// be returned.
func (c *Configuration) ParseFloat(key string, defVal float64) (float64, error) {
	val, ok := c.Props.Get(key)
	if ok {
		var err error
		result, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return defVal, fmt.Errorf("invalid float value %s=%s [%w]", key, val, err)
		}
		return result, nil
	} else {
		return defVal, nil
	}
}

// ParseByteSize converts a property value in byte size format to uint64. If
// the property does not exist, then the default value will be returned with a
// nil error. If the property value could not be parsed, then an error and the
// default value will be returned.
//
// The format supported is "<num> <suffix>" where <num> is a numeric value
// (whole number or decimal) and <suffix> is a byte size unit as listed below.
// The <suffix> and space between <num> and <suffix> are optional.
//
// The supported suffixes are:
// (none) - not modified (x 1)
// k  - kilobytes (x 1000)
// Ki - kibibytes (x 1024)
// M  - megabyte  (x 1000^2)
// Mi - mebibyte  (x 1024^2)
// G  - gigabyte  (x 1000^3)
// Gi - gibibyte  (x 1024^3)
// T  - terabyte  (x 1000^4)
// Ti - tebibyte  (x 1024^4)
// P  - petabyte  (x 1000^5)
// Pi - pebibyte  (x 1024^5)
// E  - exabyte   (x 1000^6)
// Ei - exbibyte  (x 1024^6)
func (c *Configuration) ParseByteSize(key string, defVal uint64) (uint64, error) {
	val, ok := c.Props.Get(key)
	if ok {
		match := sizePattern.FindAllStringSubmatch(val, -1)
		if match == nil || len(match) != 1 || len(match[0]) != 3 {
			return defVal, fmt.Errorf("invalid size value %s=%s", key, val)
		}
		if strings.Index(match[0][1], ".") < 0 {
			num, err := strconv.ParseUint(match[0][1], 10, 64)
			if err != nil {
				return defVal, fmt.Errorf("invalid size value %s=%s [%w]", key, val, err)
			}
			mult := c.byteSizeMult(match[0][2])
			if mult == 0 {
				return defVal, fmt.Errorf("invalid size value %s=%s (unknown suffix)", key, val)
			}
			return num * mult, nil
		} else {
			num, err := strconv.ParseFloat(match[0][1], 64)
			if err != nil {
				return defVal, fmt.Errorf("invalid size value %s=%s [%w]", key, val, err)
			}
			mult := c.byteSizeMult(match[0][2])
			if mult == 0 {
				return defVal, fmt.Errorf("invalid size value %s=%s (unknown suffix)", key, val)
			}
			return uint64(math.Round(num * float64(mult))), nil
		}
	} else {
		return defVal, nil
	}
}

// byteSizeMult determines the multiplier to be used for the given unit.
func (c *Configuration) byteSizeMult(suffix string) uint64 {
	switch suffix {
	case "":
		return 1
	case "k":
		return 1000
	case "Ki":
		return 1 << 10
	case "M":
		return 1_000_000
	case "Mi":
		return 1 << 20
	case "G":
		return 1_000_000_000
	case "Gi":
		return 1 << 30
	case "T":
		return 1_000_000_000_000
	case "Ti":
		return 1 << 40
	case "P":
		return 1_000_000_000_000_000
	case "Pi":
		return 1 << 50
	case "E":
		return 1_000_000_000_000_000_000
	case "Ei":
		return 1 << 60
	default:
		return 0
	}
}

// ParseSize converts a property value with a metric size suffix to float64. If
// the property does not exist, then the default value will be returned with a
// nil error. If the property value could not be parsed, then an error and the
// default value will be returned.
//
// The format supported is "<num> <suffix>" where <num> is a numeric value
// (whole number or decimal) and <suffix> is a size unit as listed below.
// The <suffix> and the space between <num> and <suffix> are optional.
//
// The supported suffixes are:
// Y  - yotta (10^24)
// Z  - zetta (10^21)
// E  - exa   (10^18)
// P  - peta  (10^15)
// T  - tera  (10^12)
// G  - giga  (10^9)
// M  - mega  (10^6)
// k  - kilo  (10^3)
// h  - hecto (10^2)
// da - deca  (10^1)
// (none) - not modified (x 1)
// d  - deci  (10^-1)
// c  - centi (10^-2)
// m  - milli (10^-3)
// u  - micro (10^-6)
// n  - nano  (10^-9)
// p  - pico  (10^-12)
// f  - femto (10^-15)
// a  - atto  (10^-18)
// z  - zepto (10^-21)
// y  - yocto (10^-23)
func (c *Configuration) ParseSize(key string, defVal float64) (float64, error) {
	val, ok := c.Props.Get(key)
	if ok {
		match := sizePattern.FindAllStringSubmatch(val, -1)
		if match == nil || len(match) != 1 || len(match[0]) != 3 {
			return defVal, fmt.Errorf("invalid size value %s=%s", key, val)
		}
		num, err := strconv.ParseFloat(match[0][1], 64)
		if err != nil {
			return defVal, fmt.Errorf("invalid size value %s=%s [%w]", key, val, err)
		}
		mult := c.sizeMult(match[0][2])
		if mult == 0 {
			return defVal, fmt.Errorf("invalid size value %s=%s (unknown suffix)", key, val)
		}
		return num * mult, nil
	} else {
		return defVal, nil
	}
}

// sizeMult determines the multiplier to be used for the given unit.
func (c *Configuration) sizeMult(suffix string) float64 {
	switch suffix {
	case "Y":
		return 1e24
	case "Z":
		return 1e21
	case "E":
		return 1e18
	case "P":
		return 1e15
	case "T":
		return 1e12
	case "G":
		return 1e9
	case "M":
		return 1e6
	case "k":
		return 1e3
	case "h":
		return 1e2
	case "da":
		return 1e1
	case "":
		return 1
	case "d":
		return 1e-1
	case "c":
		return 1e-2
	case "m":
		return 1e-3
	case "u":
		return 1e-6
	case "n":
		return 1e-9
	case "p":
		return 1e-12
	case "f":
		return 1e-15
	case "a":
		return 1e-18
	case "z":
		return 1e-21
	case "y":
		return 1e-24
	default:
		return 0
	}
}

// ParseBool converts a property value to a bool. If the property does not
// exist, then the default value will be returned with a nil error. If the
// property value could not be parsed, then an error and the default value will
// be returned.
//
// If the StrictBool setting is true, then only "true" and "false" values are
// able to be converted.
//
// If StrictBool is false (the default), then the following values are
// converted:
//     true, t, yes, y, 1, on -> true
//     false, f, no, n, 0, off -> false
func (c *Configuration) ParseBool(key string, defVal bool) (bool, error) {
	val, ok := c.Props.Get(key)
	if ok {
		if c.StrictBool {
			if val == "true" {
				return true, nil
			} else if val == "false" {
				return false, nil
			} else {
				return defVal, fmt.Errorf("invalid bool value %s=%s", key, val)
			}
		} else {
			val = strings.ToLower(val)
			if val == "true" || val == "t" || val == "yes" || val == "y" || val == "1" || val == "on" {
				return true, nil
			} else if val == "false" || val == "f" || val == "no" || val == "n" || val == "0" || val == "off" {
				return false, nil
			} else {
				return defVal, fmt.Errorf("invalid bool value %s=%s", key, val)
			}
		}
	} else {
		return defVal, nil
	}
}

// ParseDuration converts a property value to a Duration. If the property does
// not exist, then the default value will be returned with a nil error. If the
// property value could not be parsed, then an error and the default value will
// be returned.
//
// The format used is the same as time.ParseDuration.
func (c *Configuration) ParseDuration(key string, defVal time.Duration) (time.Duration, error) {
	val, ok := c.Props.Get(key)
	if ok {
		var err error
		result, err := time.ParseDuration(val)
		if err != nil {
			return defVal, fmt.Errorf("invalid duration value %s=%s [%w]", key, val, err)
		}
		return result, nil
	} else {
		return defVal, nil
	}
}

// ParseDate converts a property value to a Time. If the property does not
// exist, then the default value will be returned with a nil error. If the
// property value could not be parsed, then an error and the default value will
// be returned.
//
// The format used is provided by the DateFormat setting and follows the format
// defined in time.Layout. If none is set, the default of 2006-01-02 is used.
func (c *Configuration) ParseDate(key string, defVal time.Time) (time.Time, error) {
	val, ok := c.Props.Get(key)
	if ok {
		layout := c.DateFormat
		if layout == "" {
			layout = "2006-01-02"
		}
		result, err := time.Parse(layout, val)
		if err != nil {
			return defVal, fmt.Errorf("invalid date value %s=%s [%w]", key, val, err)
		}
		return result, nil
	} else {
		return defVal, nil
	}
}

// Decrypt returns the plaintext value of a property encrypted with the Encrypt
// function. If the property does not exist, then the default value will be
// returned with a nil error. If the property value could not be decrypted,
// then an error and the default value will be returned.
func (c *Configuration) Decrypt(password string, key string, defVal string) (string, error) {
	val, ok := c.Props.Get(key)
	if ok {
		dec, err := Decrypt(password, val)
		if err != nil {
			return defVal, fmt.Errorf("invalid encrypted value for %s [%w]", val, err)
		}
		return dec, nil
	} else {
		return defVal, nil
	}
}
