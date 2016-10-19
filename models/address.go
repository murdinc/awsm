package models

// Address represents an Elastic IP Address
type Address struct {
	AllocationID            string `json:"allocationID" awsmTable:"Allocation ID"`
	PublicIP                string `json:"publicIP" awsmTable:"Public IP"`
	PrivateIP               string `json:"privateIP" awsmTable:"Private IP"`
	Domain                  string `json:"domain" awsmTable:"Domain"`
	InstanceID              string `json:"instanceID" awsmTable:"Instance ID"`
	Status                  string `json:"status" awsmTable:"Status"`
	Attachment              string `json:"attachment" awsmTable:"Attachment"`
	NetworkInterfaceID      string `json:"networkInterfaceID" awsmTable:"Network Interface ID"`
	NetworkInterfaceOwnerID string `json:"networkOwnerID"`
	Region                  string `json:"region" awsmTable:"Region"`
}
