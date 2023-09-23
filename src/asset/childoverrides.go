package asset

import (
	"reflect"
	"strings"
)

type childOverrides struct {
	overrides map[string]struct{}
}

func newChildOverrides() *childOverrides {
	return &childOverrides{
		overrides: map[string]struct{}{},
	}
}

func (c *childOverrides) PathHasOverride(path string) bool {
	_, exists := c.overrides[path]
	return exists
}

func (c *childOverrides) BuildFromCommonFormat(commonFormat any) {
	c.buildFromCommonFormat(commonFormat, []string{})
}

func (c *childOverrides) RemovePath(path string) {
	delete(c.overrides, path)
}

func (c *childOverrides) Empty() bool {
	return len(c.overrides) == 0
}

func (c *childOverrides) AddPath(path string) {
	c.overrides[path] = struct{}{}
}

func (c *childOverrides) buildFromCommonFormat(commonFormat any, nameStack []string) {
	v := reflect.ValueOf(commonFormat)

	if v.Type().Kind() != reflect.Map {
		return
	}

	iter := v.MapRange()
	for iter.Next() {
		k, v := iter.Key(), iter.Value()
		if v.Elem().Kind() == reflect.Map {
			c.buildFromCommonFormat(v.Elem().Interface(), append(nameStack, k.String()))
		} else {
			c.AddPath(strings.Join(append(nameStack, k.String()), "."))
		}
	}
}
