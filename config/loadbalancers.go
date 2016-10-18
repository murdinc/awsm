package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type LoadBalancerClasses map[string]LoadBalancerClass

type LoadBalancerClass struct {
	Scheme            string                 `json:"scheme" awsmList:"Scheme"`
	SecurityGroups    []string               `json:"securityGroups" awsmList:"Security Groups"`
	Subnets           []string               `json:"subnets" awsmList:"Subnets"`
	AvailabilityZones []string               `json:"availabilityZones" awsmList:"Availability Zone"`
	Listeners         []LoadBalancerListener `json:"listeners"  awsmList:"Listeners"`
}

type LoadBalancerListener struct {
	ID               string `json:"id" hash:"ignore" awsm:"ignore"` // Needed?
	InstancePort     int    `json:"instancePort"`
	LoadBalancerPort int    `json:"loadBalancerPort"`
	Protocol         string `json:"protocol"`
	InstanceProtocol string `json:"instanceProtocol"`
	SSLCertificateID string `json:"sslCertificateID"`
}

func DefaultLoadBalancerClasses() LoadBalancerClasses {
	defaultLBs := make(LoadBalancerClasses)

	defaultLBs["prod"] = LoadBalancerClass{
		Scheme:         "",
		SecurityGroups: []string{},
		Subnets:        []string{},
		Listeners: []LoadBalancerListener{
			LoadBalancerListener{
				InstancePort:     80,
				LoadBalancerPort: 80,
				Protocol:         "tcp",
				InstanceProtocol: "tcp",
				SSLCertificateID: "",
			},
		},
		AvailabilityZones: []string{"us-west-1a"},
	}

	return defaultLBs
}

func LoadLoadBalancerClass(name string) (LoadBalancerClass, error) {
	cfgs := make(LoadBalancerClasses)
	item, err := GetItemByName("loadbalancers", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllLoadBalancerClasses() (LoadBalancerClasses, error) {
	cfgs := make(LoadBalancerClasses)
	items, err := GetItemsByType("loadbalancers")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c LoadBalancerClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "loadbalancers/", "", -1)
		cfg := new(LoadBalancerClass)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "Scheme":
				cfg.Scheme = val

			case "SecurityGroups":
				cfg.SecurityGroups = append(cfg.SecurityGroups, val)

			case "Subnets":
				cfg.Subnets = append(cfg.Subnets, val)

			case "AvailabilityZones":
				cfg.AvailabilityZones = append(cfg.AvailabilityZones, val)

			}
		}

		// Get the listeners
		listeners, _ := GetItemsByType("loadbalancers/" + name + "/listeners")
		cfg.Listeners = make([]LoadBalancerListener, len(listeners))
		for i, listener := range listeners {

			cfg.Listeners[i].ID = strings.Replace(*listener.Name, "loadbalancers/"+name+"/listeners/", "", -1) // Needed?

			for _, attribute := range listener.Attributes {

				val := *attribute.Value

				switch *attribute.Name {

				case "InstancePort":
					cfg.Listeners[i].InstancePort, _ = strconv.Atoi(val)

				case "LoadBalancerPort":
					cfg.Listeners[i].LoadBalancerPort, _ = strconv.Atoi(val)

				case "Protocol":
					cfg.Listeners[i].Protocol = val

				case "InstanceProtocol":
					cfg.Listeners[i].InstanceProtocol = val

				case "SSLCertificateID":
					cfg.Listeners[i].SSLCertificateID = val

				}
			}

		}

		c[name] = *cfg
	}
}