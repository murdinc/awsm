package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// LoadBalancerClasses is a map of Load Balancers Classes
type LoadBalancerClasses map[string]LoadBalancerClass

// LoadBalancerClass is a single Load Balancer Class
type LoadBalancerClass struct {
	Scheme            string   `json:"scheme" awsmClass:"Scheme"`
	SecurityGroups    []string `json:"securityGroups" awsmClass:"Security Groups"`
	Vpc               string   `json:"vpc" awsmClass:"VPC"`
	Subnets           []string `json:"subnets" awsmClass:"Subnets"`
	AvailabilityZones []string `json:"availabilityZones" awsmClass:"Availability Zone"`

	// Listeners
	LoadBalancerListeners []LoadBalancerListener `json:"loadBalancerListeners" hash:"ignore" awsmClass:"Listeners"`

	// Health Checks
	LoadBalancerHealthCheck LoadBalancerHealthCheck `json:"loadBalancerHealthCheck" hash:"ignore" awsmClass:"Health Check"`

	// Attributes
	LoadBalancerAttributes LoadBalancerAttributes `json:"loadBalancerAttributes" hash:"ignore" awsmClass:"Attributes"`
}

// LoadBalancerListener is a single Load Balancer Listener
type LoadBalancerListener struct {
	ID               string `json:"id" hash:"ignore" awsm:"ignore"` // Needed?
	InstancePort     int    `json:"instancePort"`
	LoadBalancerPort int    `json:"loadBalancerPort"`
	Protocol         string `json:"protocol"`
	InstanceProtocol string `json:"instanceProtocol"`
	SSLCertificateID string `json:"sslCertificateID"`
}

type LoadBalancerHealthCheck struct {
	HealthCheckTarget             string `json:"healthCheckTarget" awsmClass:"Health Check Target"`
	HealthCheckTimeout            int    `json:"healthCheckTimeout" awsmClass:"Health Check Timeout"`
	HealthCheckInterval           int    `json:"healthCheckInterval" awsmClass:"Health Check Interval"`
	HealthCheckUnhealthyThreshold int    `json:"healthCheckUnhealthyThreshold" awsmClass:"Unhealthy Threshold"`
	HealthCheckHealthyThreshold   int    `json:"healthCheckHealthyThreshold" awsmClass:"Healthy Threshold"`
}

type LoadBalancerAttributes struct {
	// Connection Draining
	ConnectionDrainingEnabled bool `json:"connectionDrainingEnabled" awsmClass:"Connection Draining"`
	ConnectionDrainingTimeout int  `json:"connectionDrainingTimeout" awsmClass:"Connection Draining Timeout"`

	// Connection Settings
	IdleTimeout                   int  `json:"idleTimeout" awsmClass:"Idle Timeout"`
	CrossZoneLoadBalancingEnabled bool `json:"crossZoneLoadBalancingEnabled" awsmClass:"Cross Zone Load Balancing"`

	// Access Logs
	AccessLogEnabled        bool   `json:"accessLogEnabled" awsmClass:"Access Log Enabled"`
	AccessLogEmitInterval   int    `json:"accessLogEmitInterval" awsmClass:"Access Log Emit Interval"`
	AccessLogS3BucketName   string `json:"accessLogS3BucketName" awsmClass:"Access Log S3 Bucket Name"`
	AccessLogS3BucketPrefix string `json:"accessLogS3BucketPrefix" awsmClass:"Access Log S3 Bucket Prefix"`
}

// DefaultLoadBalancerClasses returns the default Load Balancer Classes
func DefaultLoadBalancerClasses() LoadBalancerClasses {
	defaultLBs := make(LoadBalancerClasses)

	defaultLBs["prod"] = LoadBalancerClass{
		Scheme:         "internet-facing",
		SecurityGroups: []string{"prod"},
		Subnets:        []string{"public"},
		LoadBalancerHealthCheck: LoadBalancerHealthCheck{
			HealthCheckTarget:             "HTTP:80/index.html",
			HealthCheckTimeout:            5,
			HealthCheckInterval:           30,
			HealthCheckUnhealthyThreshold: 2,
			HealthCheckHealthyThreshold:   10,
		},
		LoadBalancerListeners: []LoadBalancerListener{
			LoadBalancerListener{
				InstancePort:     80,
				LoadBalancerPort: 80,
				Protocol:         "HTTP",
				InstanceProtocol: "HTTP",
				SSLCertificateID: "",
			},
		},
		AvailabilityZones: []string{"us-west-1a"},
	}

	return defaultLBs
}

// SaveLoadBalancerClass reads unmarshals a byte slice and inserts it into the db
func SaveLoadBalancerClass(className string, data []byte) (class LoadBalancerClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = Insert("loadbalancers", LoadBalancerClasses{className: class})
	return
}

// LoadLoadBalancerClass loads a Load Balancer Class by its name
func LoadLoadBalancerClass(name string) (LoadBalancerClass, error) {
	// awkward func name ^
	cfgs := make(LoadBalancerClasses)
	item, err := GetItemByName("loadbalancers", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllLoadBalancerClasses loads all Load Balancer Classes
func LoadAllLoadBalancerClasses() (LoadBalancerClasses, error) {
	cfgs := make(LoadBalancerClasses)
	items, err := GetItemsByType("loadbalancers")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into a Load Balancer Class
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

			case "Vpc":
				cfg.Vpc = val

			case "Subnets":
				cfg.Subnets = append(cfg.Subnets, val)

			case "AvailabilityZones":
				cfg.AvailabilityZones = append(cfg.AvailabilityZones, val)

			// Get the health check
			case "HealthCheckTarget":
				cfg.LoadBalancerHealthCheck.HealthCheckTarget = val

			case "HealthCheckTimeout":
				cfg.LoadBalancerHealthCheck.HealthCheckTimeout, _ = strconv.Atoi(val)

			case "HealthCheckInterval":
				cfg.LoadBalancerHealthCheck.HealthCheckInterval, _ = strconv.Atoi(val)

			case "HealthCheckUnhealthyThreshold":
				cfg.LoadBalancerHealthCheck.HealthCheckUnhealthyThreshold, _ = strconv.Atoi(val)

			case "HealthCheckHealthyThreshold":
				cfg.LoadBalancerHealthCheck.HealthCheckHealthyThreshold, _ = strconv.Atoi(val)

			case "ConnectionDrainingEnabled":
				cfg.LoadBalancerAttributes.ConnectionDrainingEnabled, _ = strconv.ParseBool(val)

			// Get the attributes
			case "ConnectionDrainingTimeout":
				cfg.LoadBalancerAttributes.ConnectionDrainingTimeout, _ = strconv.Atoi(val)

			case "IdleTimeout":
				cfg.LoadBalancerAttributes.IdleTimeout, _ = strconv.Atoi(val)

			case "CrossZoneLoadBalancingEnabled":
				cfg.LoadBalancerAttributes.CrossZoneLoadBalancingEnabled, _ = strconv.ParseBool(val)

			case "AccessLogEnabled":
				cfg.LoadBalancerAttributes.AccessLogEnabled, _ = strconv.ParseBool(val)

			case "AccessLogEmitInterval":
				cfg.LoadBalancerAttributes.AccessLogEmitInterval, _ = strconv.Atoi(val)

			case "AccessLogS3BucketName":
				cfg.LoadBalancerAttributes.AccessLogS3BucketName = val

			case "AccessLogS3BucketPrefix":
				cfg.LoadBalancerAttributes.AccessLogS3BucketPrefix = val

			}
		}

		// Get the listeners
		listeners, _ := GetItemsByType("loadbalancers/" + name + "/listeners")
		cfg.LoadBalancerListeners = make([]LoadBalancerListener, len(listeners))
		for i, listener := range listeners {

			cfg.LoadBalancerListeners[i].ID = strings.Replace(*listener.Name, "loadbalancers/"+name+"/listeners/", "", -1) // Needed?

			for _, attribute := range listener.Attributes {

				val := *attribute.Value

				switch *attribute.Name {

				case "InstancePort":
					cfg.LoadBalancerListeners[i].InstancePort, _ = strconv.Atoi(val)

				case "LoadBalancerPort":
					cfg.LoadBalancerListeners[i].LoadBalancerPort, _ = strconv.Atoi(val)

				case "Protocol":
					cfg.LoadBalancerListeners[i].Protocol = val

				case "InstanceProtocol":
					cfg.LoadBalancerListeners[i].InstanceProtocol = val

				case "SSLCertificateID":
					cfg.LoadBalancerListeners[i].SSLCertificateID = val

				}
			}

		}

		c[name] = *cfg
	}
}
