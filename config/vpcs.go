package config

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// VpcClasses is a map of Vpc Classes
type VpcClasses map[string]VpcClass

// VpcClass is a single Vpc Class
type VpcClass struct {
	CIDR    string `json:"cidr" awsmClass:"CIDR"`
	Tenancy string `json:"tenancy" awsmClass:"Tenancy"`
}

// DefaultVpcClasses returns the default Vpc Classes
func DefaultVpcClasses() VpcClasses {
	defaultVpcs := make(VpcClasses)

	defaultVpcs["awsm"] = VpcClass{
		CIDR:    "/16",
		Tenancy: "default",
	}

	return defaultVpcs
}

// SaveVpcClass reads unmarshals a byte slice and inserts it into the db
func SaveVpcClass(className string, data []byte) (class VpcClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = Insert("vpcs", VpcClasses{className: class})
	return
}

// LoadVpcClass loads a Vpc Class by its name
func LoadVpcClass(name string) (VpcClass, error) {
	cfgs := make(VpcClasses)
	item, err := GetItemByName("vpcs", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllVpcClasses loads all Vpc Classes
func LoadAllVpcClasses() (VpcClasses, error) {
	cfgs := make(VpcClasses)
	items, err := GetItemsByType("vpcs")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into a Vpc Class
func (c VpcClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "vpcs/", "", -1)
		cfg := new(VpcClass)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "CIDR":
				cfg.CIDR = val

			case "Tenancy":
				cfg.Tenancy = val

			}
		}

		c[name] = *cfg
	}
}
