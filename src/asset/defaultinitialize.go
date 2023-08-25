package asset

import (
	"reflect"
)

func callAllDefaultInitializers(obj any) {
	if obj == nil {
		return
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	//fmt.Printf("%#v\n", v)
	if v.IsNil() {
		return
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}
	processValue := func(v reflect.Value) {
		if v.Kind() == reflect.Interface {
			callAllDefaultInitializers(v.Interface())
		} else if v.Kind() == reflect.Pointer {
			callAllDefaultInitializers(v.Interface())
		} else if v.CanAddr() {
			callAllDefaultInitializers(v.Addr().Interface())
		} else {
			callAllDefaultInitializers(v.Interface())
		}
	}

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			tField := t.Field(i)
			if tField.IsExported() {
				vField := v.Field(i)
				processValue(vField)
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			processValue(v.Index(i))
		}
	}

	if di, ok := obj.(DefaultInitializer); ok {
		di.DefaultInitialize()
	}
}
