package models

// KeyPair represents an SSH KeyPair
type KeyPair struct {
	KeyName        string `json:"keyName" awsmTable:"Key Name"`
	KeyFingerprint string `json:"keyFingerprint" awsmTable:"Key Fingerprint"`
	Region         string `json:"region" awsmTable:"Region"`
}
