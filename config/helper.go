package config

import (
	"fmt"
	"reflect"
	"strings"
)

func ExtractAwsmList(in interface{}) (keys, values []string) {

	t := reflect.TypeOf(in)
	tV := reflect.ValueOf(in)

	value := reflect.New(t).Interface()

	v := reflect.ValueOf(value)
	i := reflect.Indirect(v)
	s := i.Type()
	fields := s.NumField()

	for k := 0; k < fields; k++ {
		sTag := t.Field(k).Tag.Get("awsmList")

		var sVal string

		switch tV.Field(k).Interface().(type) {
		case int:
			sVal = fmt.Sprint(tV.Field(k).Int())
		case string:
			sVal = tV.Field(k).String()
		case bool:
			sVal = fmt.Sprint(tV.Field(k).Bool())
		case []string:
			sVal = strings.Join(tV.Field(k).Interface().([]string), ", ")

			// TODO other types?
		}

		if sTag != "" {
			keys = append(keys, sTag)
			values = append(values, sVal)
		}
	}

	return
}
