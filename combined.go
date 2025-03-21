// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

// Combined provides property value lookups across multiple sources.
type Combined struct {
	// The property sources to use for lookup in priority order. The first
	// source to have a value for a property will be used.
	Sources []PropertyGetter
}

// Ensure that Combined implements PropertyGetter
var _ PropertyGetter = &Combined{}

// Get retrieves the value of a property from the source list. If the source
// list is empty or none of the sources has the property, an empty string will
// be returned. The bool return value indicates whether the property was found.
func (c *Combined) Get(key string) (string, bool) {
	if c.Sources == nil {
		return "", false
	}
	for _, l := range c.Sources {
		val, ok := l.Get(key)
		if ok {
			return val, true
		}
	}
	return "", false
}

// GetDefault retrieves the value of a property from the source list. If the
// source list is empty or none of the sources has the property, then the
// default value will be returned.
func (c *Combined) GetDefault(key string, defVal string) string {
	val, ok := c.Get(key)
	if ok {
		return val
	} else {
		return defVal
	}
}

// Names returns the unique names of all properties that have been set.
func (c *Combined) Names() []string {
	vals := make(map[string]struct{})
	for _, l := range c.Sources {
		for _, v := range l.Names() {
			vals[v] = struct{}{}
		}
	}

	result := make([]string, 0, len(vals))
	for k := range vals {
		result = append(result, k)
	}
	return result
}
