package models

// HostedZone represents a Route53 HostedZone
type HostedZone struct {
	Name                   string `json:"name" awsmTable:"Name"`
	PrivateZone            bool   `json:"private" awsmTable:"Private"`
	ResourceRecordSetCount int    `json:"resourceRecordSetCount" awsmTable:"Record Count"`
	Id                     string `json:"id" awsmTable:"Id"`
	Comment                string `json:"comment"`
}

// ResourceRecord represents a Route53 Resource Record
type ResourceRecord struct {
	Name          string      `json:"name" awsmTable:"Name"`
	Type          string      `json:"type" awsmTable:"Type"`
	TTL           int         `json:"ttl" awsmTable:"TTL"`
	HealthCheckId string      `json:"region"`
	Values        []string    `json:"values"`
	TableValues   []string    `json:"tableValues" awsmTable:"Values"`
	AliasTarget   AliasTarget `json:"aliasTarget"`
	Region        string      `json:"region" awsmTable:"Region"`
	Failover      string      `json:"region" awsmTable:"Failover"`
	HostedZoneId  string      `json:"hostedZoneId" awsmTable:"Hosted Zone Id"`
	//GeoLocation
}

// ResourceRecordChange represents a Route53 Resource Record Change
type ResourceRecordChange struct {
	Action string   `json:"action" awsmTable:"Action"`
	Name   string   `json:"name" awsmTable:"Name"`
	Values []string `json:"values" awsmTable:"Values"`
	Type   string   `json:"type" awsmTable:"Type"`
	TTL    int      `json:"ttl" awsmTable:"TTL"`
}

type AliasTarget struct {
	DNSName              string
	EvaluateTargetHealth bool
	HostedZoneId         string
}
