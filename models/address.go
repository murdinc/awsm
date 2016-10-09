package models

type Address struct {
	AllocationId            string `json:"allocationId" awsmTable:"Allocation ID"`
	PublicIp                string `json:"publicIp" awsmTable:"Public IP"`
	PrivateIp               string `json:"privateIp" awsmTable:"Private IP"`
	Domain                  string `json:"domain" awsmTable:"Domain"`
	InstanceId              string `json:"instanceId" awsmTable:"Instance ID"`
	Status                  string `json:"status" awsmTable:"Status"`
	Attachment              string `json:"attachment" awsmTable:"Attachment"`
	NetworkInterfaceId      string `json:"networkInterfaceId" awsmTable:"Network Interface ID"`
	NetworkInterfaceOwnerId string `json:"networkOwnerId" awsmTable:"Network Interface Owner ID"`
	Region                  string `json:"region" awsmTable:"Region"`
}
