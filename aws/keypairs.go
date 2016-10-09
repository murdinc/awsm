package aws

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"reflect"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
)

type KeyPairs []KeyPair

type KeyPair models.KeyPair

func GetKeyPairByName(region, name string) (KeyPair, error) {

	if len(name) < 1 {
		return KeyPair{}, errors.New("No KeyName provided!")
	}

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{
			aws.String(name),
		},
	}
	result, err := svc.DescribeKeyPairs(params)

	if err != nil {
		return KeyPair{}, err
	}

	count := len(result.KeyPairs)

	switch count {
	case 0:
		return KeyPair{}, errors.New("No KeyPair found named [" + name + "] in [" + region + "]!")
	case 1:
		keyPair := new(KeyPair)
		keyPair.Marshal(result.KeyPairs[0], region)
		return *keyPair, nil
	}

	return KeyPair{}, errors.New("Found more than one KeyPair named [" + name + "] in [" + region + "]!")

}

func GetKeyPairs(search string) (*KeyPairs, []error) {
	var wg sync.WaitGroup
	var errs []error

	keyList := new(KeyPairs)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionKeyPairs(*region.RegionName, keyList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering key pair list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return keyList, errs
}

func (k *KeyPair) Marshal(keyPair *ec2.KeyPairInfo, region string) {
	k.KeyName = aws.StringValue(keyPair.KeyName)
	k.KeyFingerprint = aws.StringValue(keyPair.KeyFingerprint)
	k.Region = region
}

func GetRegionKeyPairs(region string, keyList *KeyPairs, search string) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		return err
	}

	key := make(KeyPairs, len(result.KeyPairs))
	for i, keyPair := range result.KeyPairs {
		key[i].Marshal(keyPair, region)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, kp := range key {
			rKey := reflect.ValueOf(kp)

			for k := 0; k < rKey.NumField(); k++ {
				sVal := rKey.Field(k).String()

				if term.MatchString(sVal) {
					*keyList = append(*keyList, key[i])
					continue Loop
				}
			}
		}
	} else {
		*keyList = append(*keyList, key[:]...)
	}

	return nil
}

func (i *KeyPairs) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.KeyName,
			val.KeyFingerprint,
			val.Region,
		}
	}

	table.SetHeader([]string{"Key Name", "Key Fingerprint", "Region"})

	table.AppendBulk(rows)
	table.Render()
}

func CreateAndImportKeyPair(name string, dryRun bool) []error {

	var wg sync.WaitGroup
	var errs []error

	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Create the KeyPair locally
	pubKey, err := MakeKeyPair(name, dryRun)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	// Import the KeyPair to each AWS Region
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()

			err := ImportKeyPair(*region.RegionName, name, pubKey, dryRun)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error importing public key to region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return errs
}

func ImportKeyPair(region, name string, publicKey []byte, dryRun bool) error {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(name),
		PublicKeyMaterial: []byte(publicKey),
		DryRun:            aws.Bool(dryRun),
	}
	_, err := svc.ImportKeyPair(params)

	if err != nil {
		return err
	}

	terminal.Information("Imported public key for [" + name + "] into [" + region + "]!")

	return nil
}

func DeleteKeyPairs(name string, dryRun bool) (errs []error) {

	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	keyList, err := GetKeyPairs(name)
	if err != nil {
		terminal.ErrorLine("Error gathering KeyPair list")
		errs = append(errs, err...)
		return
	}

	if len(*keyList) > 0 {
		// Print the table
		keyList.PrintTable()
	} else {
		terminal.ErrorLine("No KeyPairs found, Aborting!")
		return
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these KeyPairs?") {
		terminal.ErrorLine("Aborting!")
		return
	}

	// Delete 'Em
	for _, key := range *keyList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(key.Region)}))

		params := &ec2.DeleteKeyPairInput{
			KeyName: aws.String(key.KeyName),
			DryRun:  aws.Bool(dryRun),
		}

		_, err := svc.DeleteKeyPair(params)

		if err != nil {
			terminal.ErrorLine(err.Error())
			errs = append(errs, err)
		} else {
			terminal.Information("Deleted KeyPair [" + key.KeyName + "] in region [" + key.Region + "]!")
		}

	}

	return
}

// MakeSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
// http://stackoverflow.com/a/34347463
func MakeKeyPair(name string, dryRun bool) ([]byte, error) {

	var publicKey []byte

	currentUser, _ := user.Current()
	sshLocation := currentUser.HomeDir + "/.ssh/"

	privateKeyPath := sshLocation + name + ".pem"
	publicKeyPath := sshLocation + name + ".pub"

	if !dryRun {

		// Check that the key doesn't exist yet
		if _, err := os.Stat(privateKeyPath); !os.IsNotExist(err) {
			terminal.Information("Local key named [" + name + "] already exists, reading from existing public key file!")

			publicKey, err = ioutil.ReadFile(publicKeyPath)
			if err != nil {
				terminal.ErrorLine("Error while reading public key file!")
				return publicKey, err
			}

		} else {
			privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
			if err != nil {
				return publicKey, err
			}

			// Generate and write private key as PEM
			privateKeyFile, err := os.Create(privateKeyPath)
			defer privateKeyFile.Close()
			if err != nil {
				return publicKey, err
			}

			privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
			if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
				return publicKey, err
			}

			// Generate and write public key
			pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
			if err != nil {
				return publicKey, err
			}

			publicKey = ssh.MarshalAuthorizedKey(pub)

			err = ioutil.WriteFile(publicKeyPath, publicKey, 0655)
			if err != nil {
				return publicKey, err
			}

			terminal.Information("Created KeyPair named [" + name + "]")
		}

	} else {
		publicKey = []byte("dryRun!")
	}

	return publicKey, nil
}
