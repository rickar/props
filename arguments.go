// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"os"
	"strings"
)

// Arguments reads properties from the command line arguments.
// Property arguments are expected to have a common prefix and use key=value
// format. Other arguments are ignored.
//
// For example, the command:
//   cmd -a -1 -z --prop.1=a --prop.2=b --prop.3 --log=debug
// with a prefix of '--prop.' would have properties "1"="a", "2"="b", and
// "3"="".
type Arguments struct {
	// Prefix provides the common prefix to use when looking for property
	// arguments. If not set, the default of '--' will be used.
	Prefix string
}

// Ensure that Arguments implements PropertyGetter
var _ PropertyGetter = &Arguments{}

// Get retrieves the value of a property from the command line arguments. If
// the property does not exist, an empty string will be returned. The bool
// return value indicates whether the property was found.
func (a *Arguments) Get(key string) (string, bool) {
	prefix := a.Prefix
	if prefix == "" {
		prefix = "--"
	}
	prefix = prefix + key + "="
	for _, val := range os.Args {
		if strings.HasPrefix(val, prefix) {
			return val[len(prefix):], true
		}
	}
	return "", false
}

// GetDefault retrieves the value of a property from the command line arguments.
// If the property does not exist, then the default value will be returned.
func (e *Arguments) GetDefault(key, defVal string) string {
	v, ok := e.Get(key)
	if !ok {
		return defVal
	}
	return v
}
