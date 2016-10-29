package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/terminal"
	"golang.org/x/crypto/ssh"
)

// KeyPairClasses is a map of Image classes
type KeyPairClasses map[string]KeyPairClass

// KeyPairClass is a single Image class
type KeyPairClass struct {
	Description string `json:"description" awsmClass:"Description"`
	PublicKey   string `json:"publicKey" awsmClass:"Public Key"`
	PrivateKey1 string `json:"-"`
	PrivateKey2 string `json:"-"`
	PrivateKey3 string `json:"-"`
	PrivateKey4 string `json:"-"`
}

// DefaultKeyPairClasses returns the default Image classes
func DefaultKeyPairClasses() KeyPairClasses {
	defaultKeyPairs := make(KeyPairClasses)
	defaultKeyNames := []string{"awsm"}

	for _, keyName := range defaultKeyNames {
		publicKey, privateKey, err := GenerateKeyPair(keyName)
		if err != nil {
			terminal.ErrorLine("Error while generating keypair class: " + keyName)
			return defaultKeyPairs
		}

		privateKeyLen := len(privateKey) / 4

		defaultKeyPairs[keyName] = KeyPairClass{
			Description: "Default KeyPair",
			PublicKey:   publicKey,
			PrivateKey1: privateKey[:privateKeyLen],
			PrivateKey2: privateKey[privateKeyLen : privateKeyLen*2],
			PrivateKey3: privateKey[privateKeyLen*2 : privateKeyLen*3],
			PrivateKey4: privateKey[privateKeyLen*3:],
		}
	}

	return defaultKeyPairs
}

// SaveKeyPairClass reads unmarshals a byte slice and inserts it into the db
func SaveKeyPairClass(className string, data []byte) (class KeyPairClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = InsertClasses("keypairs", KeyPairClasses{className: class})

	if err != nil {
		println(err)
	}
	return
}

// LoadKeyPairClass returns a single Image class by its name
func LoadKeyPairClass(name string) (KeyPairClass, error) {
	cfgs := make(KeyPairClasses)
	item, err := GetItemByName("keypairs", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllKeyPairClasses returns all Image classes
func LoadAllKeyPairClasses() (KeyPairClasses, error) {
	cfgs := make(KeyPairClasses)
	items, err := GetItemsByType("keypairs")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into Image Classes
func (c KeyPairClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "keypairs/", "", -1)
		cfg := new(KeyPairClass)

		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "Description":
				cfg.Description = val

			case "PublicKey":
				cfg.PublicKey = val

			case "PrivateKey1":
				cfg.PrivateKey1 = val

			case "PrivateKey2":
				cfg.PrivateKey2 = val

			case "PrivateKey3":
				cfg.PrivateKey3 = val

			case "PrivateKey4":
				cfg.PrivateKey4 = val

			}
		}
		c[name] = *cfg
	}
}

// GenerateKeyPair creates an ssh keypair
func GenerateKeyPair(name string) (publicKeyEncoded string, privateKeyEncoded string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	privateKeyEncoded = string(pem.EncodeToMemory(privateKeyPEM))

	// Generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return
	}

	publicKeyEncoded = string(ssh.MarshalAuthorizedKey(pub))

	return
}
