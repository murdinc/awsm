package config

import (
	"fmt"
	"reflect"
)

// ExtractAwsmClass extracts the tagged keys and values from an awsm class config struct for displaying on the frontend
func ExtractAwsmClass(in interface{}) (keys, values []string) {

	inType := reflect.TypeOf(in)
	inValue := reflect.ValueOf(in)

	emtpyStruct := reflect.New(inType).Interface()

	emtpyStructValue := reflect.ValueOf(emtpyStruct)
	emtpyStructIndirect := reflect.Indirect(emtpyStructValue)
	fields := emtpyStructIndirect.Type().NumField()

	for k := 0; k < fields; k++ {
		sTag := inType.Field(k).Tag.Get("awsmClass")

		if sTag != "" {

			var sVal string

			switch inValue.Field(k).Type().String() {
			case "int":
				sVal = fmt.Sprint(inValue.Field(k).Int())
			case "string":
				sVal = inValue.Field(k).String()
			case "bool":
				sVal = fmt.Sprint(inValue.Field(k).Bool())
			case "[]string":

				//println(inValue.Field(k))

				fmt.Printf(">  %#v", inValue.Field(k).String())
				/*
					vals := inValue.Field(k). .Interface()

					sVal = strings.Join(
						vals.([]string),
						", ",
					)
				*/
			default:
				fmt.Printf("ExtractAwsmClass does not have a switch for type: %#v\n", inValue.Field(k).Type().String())

			}

			keys = append(keys, sTag)
			values = append(values, sVal)
		}
	}

	return
}
