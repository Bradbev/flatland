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

func (c *childOverrides) addPath(path string) {
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
			c.addPath(strings.Join(append(nameStack, k.String()), "."))
		}
	}
}
