package config

import (
	"fmt"
	"reflect"
	"strings"
)

// ExtractAwsmClass extracts the tagged keys and values from an awsm class config struct for displaying on the frontend
func ExtractAwsmClass(in interface{}) (keys, values []string) {

	t := reflect.TypeOf(in)
	tV := reflect.ValueOf(in)

	value := reflect.New(t).Interface()

	v := reflect.ValueOf(value)
	i := reflect.Indirect(v)
	s := i.Type()
	fields := s.NumField()

	for k := 0; k < fields; k++ {
		sTag := t.Field(k).Tag.Get("awsmClass")

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
