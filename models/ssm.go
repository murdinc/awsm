package models

import "time"

// CommandInvocation represents a SSM Command Invocation
type CommandInvocation struct {
	CommandID         string          `json:"commandID" awsmTable:"Command ID"`
	Status            string          `json:"status" awsmTable:"Status"`
	StatusDetails     string          `json:"statusDetails"  awsmTable:"Status Details"`
	InstanceName      string          `json:"instanceName" awsmTable:"Instance Name"`
	InstanceID        string          `json:"instanceID" awsmTable:"Instance ID"`
	DocumentName      string          `json:"documentName" awsmTable:"Document Name"`
	Comment           string          `json:"comment"`
	RequestedDateTime time.Time       `json:"requestedDateTime" awsmTable:"Requested"`
	ServiceRole       string          `json:"serviceRole" awsmTable:"Service Role"`
	Region            string          `json:"region" awsmTable:"Region"`
	CommandPlugins    []CommandPlugin `json:"commandPlugins"`
}

// CommandPlugin represents a SSM Command Plugin
type CommandPlugin struct {
	Name                   string    `json:"name" awsmTable:"Name"`
	Output                 string    `json:"output" awsmTable:"Output"`
	ResponseCode           int       `json:"responseCode" awsmTable:"Response Code"`
	ResponseStartDateTime  time.Time `json:"responseStartDateTime" awsmTable:"Response Start"`
	ResponseFinishDateTime time.Time `json:"responseFinishDateTime" awsmTable:"Response Finish"`
}

type Entity struct {
	InstanceName    string    `json:"name" awsmTable:"Name"`
	InstanceClass   string    `json:"class" awsmTable:"Class"`
	InstanceID      string    `json:"instanceId"`
	IpAddress       string    `json:"IP Address" awsmTable:"IP Address"`
	EntityID        string    `json:"entityID" awsmTable:"Entity ID"`
	ComputerName    string    `json:"computerName" awsmTable:"Computer Name"`
	CaptureTime     time.Time `json:"captureTime" awsmTable:"Capture Time"`
	PlatformName    string    `json:"platformName" awsmTable:"Platform Name"`
	PlatformType    string    `json:"platformType" awsmTable:"Platform Type"`
	PlatformVersion string    `json:"platformVersion" awsmTable:"Platform Version"`
	ResourceType    string    `json:"resourceType" awsmTable:"Resource Type"`
	AgentType       string    `json:"agentType" awsmTable:"Agent Type"`
	AgentVersion    string    `json:"agentVersion" awsmTable:"Agent Version"`
	ContentHash     string    `json:"contentHash"`
	TypeName        string    `json:"typeName"`
	SchemaVersion   string    `json:"schemaVersion"`
	Region          string    `json:"region"`
}

type SSMInstance struct {
	Name             string    `json:"Name" awsmTable:"Name"`
	Class            string    `json:"Class" awsmTable:"Class"`
	ComputerName     string    `json:"computerName" awsmTable:"Computer Name"`
	InstanceID       string    `json:"instanceId" awsmTable:"Instance ID"`
	IPAddress        string    `json:"ipAddress" awsmTable:"IP Address"`
	LastPingDateTime time.Time `json:"lastPingDateTime" awsmTable:"Last Ping"`
	PingStatus       string    `json:"pingStatus" awsmTable:"Ping Status"`
	PlatformName     string    `json:"platformName" awsmTable:"Platform Name"`
	PlatformType     string    `json:"platformType" awsmTable:"Platform Type"`
	PlatformVersion  string    `json:"platformVersion" awsmTable:"Platform Version"`
	AgentVersion     string    `json:"agentVersion" awsmTable:"Agent Version"`
	IsLatestVersion  bool      `json:"isLatestVersion" awsmTable:"Latest Version"`
	ResourceType     string    `json:"resourceType" awsmTable:"Resource Type"`
	Region           string    `json:"region" awsmTable:"Region"`
}
