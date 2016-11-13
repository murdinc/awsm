package models

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// ExtractAwsmTable Extracts the keys and values of a struct for use in building tables of assets
func ExtractAwsmTableLinks(index int, in interface{}, header *[]string, rows *[][]string, links *[]map[string]string) {

	t := reflect.TypeOf(in)
	tV := reflect.ValueOf(in)

	value := reflect.New(t).Interface()

	v := reflect.ValueOf(value)
	i := reflect.Indirect(v)
	s := i.Type()
	fields := s.NumField()

	// Make out links map for this row
	(*links)[index] = make(map[string]string)

	for k := 0; k < fields; k++ {
		sTag := t.Field(k).Tag.Get("awsmTable")

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

		case time.Time:
			sVal = humanize.Time(tV.Field(k).Interface().(time.Time))

		default:
			println("ExtractAwsmTable does not have a switch for type:")
			println(tV.Field(k).Type().String())

			// TODO other types?
		}

		if sTag != "" {
			// Head
			if index == 0 {
				*header = append(*header, sTag)
			}
			// Rows
			(*rows)[index] = append((*rows)[index], sVal)

		}

		lTag := t.Field(k).Tag.Get("awsmLink")
		if lTag != "" {
			// Links
			(*links)[index][lTag] = sVal

		}
	}
}

// ExtractAwsmTable Extracts the keys and values of a struct for use in building tables of assets
func ExtractAwsmTable(index int, in interface{}, header *[]string, rows *[][]string) {

	t := reflect.TypeOf(in)
	tV := reflect.ValueOf(in)

	value := reflect.New(t).Interface()

	v := reflect.ValueOf(value)
	i := reflect.Indirect(v)
	s := i.Type()
	fields := s.NumField()

	for k := 0; k < fields; k++ {
		sTag := t.Field(k).Tag.Get("awsmTable")

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

		case time.Time:
			sVal = humanize.Time(tV.Field(k).Interface().(time.Time))

		default:
			println("ExtractAwsmTable does not have a switch for type:")
			println(tV.Field(k).Type().String())

			// TODO other types?
		}

		if sTag != "" {
			// Head
			if index == 0 {
				*header = append(*header, sTag)
			}
			// Rows
			(*rows)[index] = append((*rows)[index], sVal)
		}
	}
}
