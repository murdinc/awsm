package config

import (
	"fmt"
	"reflect"
	"strings"
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
			case "float64":
				sVal = fmt.Sprint(inValue.Field(k).Float())
			case "string":
				sVal = inValue.Field(k).String()
			case "bool":
				sVal = fmt.Sprint(inValue.Field(k).Bool())
			case "[]string":
				if !inValue.Field(k).IsNil() {
					strSlice, ok := inValue.Field(k).Interface().([]string)
					if ok {
						sVal = strings.Join(strSlice, ", ")
					}
				}

			case "[]config.SecurityGroupGrant":
				grants := inValue.Field(k).Interface().([]SecurityGroupGrant)
				for _, grant := range grants {

					direction := ">"
					if grant.Type == "egress" {
						direction = "<"
					}

					sVal += fmt.Sprintf("%s:%d%s:%d\n\n", grant.IPProtocol, grant.FromPort, direction, grant.ToPort)
				}

			case "[]config.LoadBalancerListener":
				listeners := inValue.Field(k).Interface().([]LoadBalancerListener)
				for _, listener := range listeners {
					sVal += fmt.Sprintf("%s:%d>%s:%d\n\n", listener.Protocol, listener.LoadBalancerPort, listener.InstanceProtocol, listener.InstancePort)
				}

			default:
				fmt.Printf("ExtractAwsmClass does not have a switch for type: %#v\n", inValue.Field(k).Type().String())

			}

			keys = append(keys, sTag)
			values = append(values, sVal)
		}
	}

	return
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return x == nil || x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

// ExtractAwsmWidget extracts the tagged keys and values from an awsm class config struct for displaying on the frontend
func ExtractAwsmWidget(in interface{}) (keys, values []string) {

	inType := reflect.TypeOf(in)
	inValue := reflect.ValueOf(in)

	emtpyStruct := reflect.New(inType).Interface()

	emtpyStructValue := reflect.ValueOf(emtpyStruct)
	emtpyStructIndirect := reflect.Indirect(emtpyStructValue)
	fields := emtpyStructIndirect.Type().NumField()

	for k := 0; k < fields; k++ {
		sTag := inType.Field(k).Tag.Get("awsmWidget")

		if sTag != "" {

			var sVal string

			switch inValue.Field(k).Type().String() {
			case "int":
				sVal = fmt.Sprint(inValue.Field(k).Int())
			case "float64":
				sVal = fmt.Sprint(inValue.Field(k).Float())
			case "string":
				sVal = inValue.Field(k).String()
			case "bool":
				sVal = fmt.Sprint(inValue.Field(k).Bool())
			case "[]string":
				sVal = strings.Join(inValue.Field(k).Interface().([]string), ", ")

			case "[]config.SecurityGroupGrant":
				grants := inValue.Field(k).Interface().([]SecurityGroupGrant)
				for _, grant := range grants {

					direction := ">"
					if grant.Type == "egress" {
						direction = "<"
					}

					sVal += fmt.Sprintf("%s:%d%s:%d\n\n", grant.IPProtocol, grant.FromPort, direction, grant.ToPort)
				}

			case "[]config.LoadBalancerListener":
				listeners := inValue.Field(k).Interface().([]LoadBalancerListener)
				for _, listener := range listeners {
					sVal += fmt.Sprintf("%s:%d>%s:%d\n\n", listener.Protocol, listener.LoadBalancerPort, listener.InstanceProtocol, listener.InstancePort)
				}

			default:
				fmt.Printf("ExtractAwsmClass does not have a switch for type: %#v\n", inValue.Field(k).Type().String())

			}

			keys = append(keys, sTag)
			values = append(values, sVal)
		}
	}

	return
}
