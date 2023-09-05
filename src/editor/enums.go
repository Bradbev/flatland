package editor

import (
	"reflect"
	"sort"
)

type enumInfo struct {
	enums map[reflect.Type]map[int32]string
}

// mapping has any as a key so that creating the map needs no casting.
// Internally, the keys of mapping will be convered to int32
func (e *enumInfo) RegisterEnum(typ reflect.Type, mapping map[any]string) {
	m := map[int32]string{}
	for k, v := range mapping {
		i32 := int32(reflect.ValueOf(k).Int())
		m[i32] = v
	}
	e.enums[typ] = m
}

func (e *enumInfo) IsKnown(typ reflect.Type) bool {
	_, exists := e.enums[typ]
	return exists
}

func (e *enumInfo) Strings(typ reflect.Type) []string {
	if mapping, ok := e.enums[typ]; ok {
		result := []string{}
		for _, s := range mapping {
			result = append(result, s)
		}
		sort.Strings(result)
		return result
	}
	return nil
}

func (e *enumInfo) StringToValue(typ reflect.Type, s string) int32 {
	if mapping, ok := e.enums[typ]; ok {
		for v, name := range mapping {
			if name == s {
				return v
			}
		}
	}
	return 0
}

func (e *enumInfo) ValueToString(typ reflect.Type, value int32) string {
	if mapping, ok := e.enums[typ]; ok {
		if v, ok := mapping[value]; ok {
			return v
		}
	}
	return ""
}
