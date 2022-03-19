// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"os"
	"strings"
)

// Environment reads properties from the OS environment.
type Environment struct {
	// Normalize indicates that key values should be converted to POSIX-style
	// environment variable names.
	//
	// If true, key values passed to Get and GetDefault will be converted to:
	// - Uppercase
	// - Alphanumeric (as per ASCII)
	// - All non-alphanumeric characters replaced with underscore '_'
	//
	// For example, 'foo.bar.baz' would become 'FOO_BAR_BAZ' and
	// '$my-test#val_1' would become '_MY_TEST_VAL_1'.
	Normalize bool
}

// Ensure that Environment implements PropertyGetter
var _ PropertyGetter = &Environment{}

// Get retrieves the value of a property from the environment. If the env var
// does not exist, an empty string will be returned. The bool return value
// indicates whether the property was found.
func (e *Environment) Get(key string) (string, bool) {
	if e.Normalize {
		envKey := strings.Map(normalizeEnv, key)
		return os.LookupEnv(envKey)
	} else {
		return os.LookupEnv(key)
	}
}

// GetDefault retrieves the value of a property from the environment. If the
// env var does not exist, then the default value will be returned.
func (e *Environment) GetDefault(key, defVal string) string {
	v, ok := e.Get(key)
	if !ok {
		return defVal
	}
	return v
}

// normalizeEnv converts a rune into a suitable replacement for an environment
// variable name.
func normalizeEnv(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	} else if r >= 'A' && r <= 'Z' {
		return r
	} else if r >= '0' && r <= '9' {
		return r
	} else {
		return '_'
	}
}
