package aws

import (
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
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// KeyPairs represents a slice of AWS KeyPairs
type KeyPairs []KeyPair

// KeyPair represents a single AWS KeyPair
type KeyPair models.KeyPair

// GetKeyPairByName returns a single KeyPair given the provided region and name
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

// GetKeyPairs returns a slice of KeyPairs that match the provided search term
func GetKeyPairs(search string) (*KeyPairs, []error) {
	var wg sync.WaitGroup
	var errs []error

	keyList := new(KeyPairs)
	regions := regions.GetRegionList()

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

// Marshal parses the response from the aws sdk into an awsm KeyPair
func (k *KeyPair) Marshal(keyPair *ec2.KeyPairInfo, region string) {
	k.KeyName = aws.StringValue(keyPair.KeyName)
	k.KeyFingerprint = aws.StringValue(keyPair.KeyFingerprint)
	k.Region = region
}

// GetRegionKeyPairs returns a list of KeyPairs for a given region into the provided KeyPairs slice
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

// PrintTable Prints an ascii table of the list of KeyPairs
func (i *KeyPairs) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No KeyPairs Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, key := range *i {
		models.ExtractAwsmTable(index, key, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// CreateKeyPair creates a KeyPair of a specified class in the specified region
func CreateKeyPair(class, region string, dryRun bool) error {

	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// KeyPair Class Config
	keypairCfg, err := config.LoadKeyPairClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found KeyPair class configuration for [" + class + "]!")

	// Validate the region
	if !regions.ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// Import the KeyPair to the requested region
	err = importKeyPair(region, class, []byte(keypairCfg.PublicKey), dryRun)
	if err != nil {
		return err
	}

	return nil
}

func importKeyPair(region, name string, publicKey []byte, dryRun bool) error {

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

	terminal.Delta("Imported public key for [" + name + "] into [" + region + "]!")

	return nil
}

// DeleteKeyPairs deletes an existing KeyPair from AWS
func DeleteKeyPairs(name string, dryRun bool) error {

	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	keyList, err := GetKeyPairs(name)
	if err != nil {
		terminal.ErrorLine("Error gathering KeyPair list")
		return nil
	}

	if len(*keyList) > 0 {
		// Print the table
		keyList.PrintTable()
	} else {
		terminal.ErrorLine("No KeyPairs found, Aborting!")
		return nil
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these KeyPairs?") {
		terminal.ErrorLine("Aborting!")
		return nil
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
		} else {
			terminal.Delta("Deleted KeyPair [" + key.KeyName + "] in region [" + key.Region + "]!")
		}

	}

	return nil
}

// InstallKeyPair installs a keypair locally
func InstallKeyPair(class string, dryRun bool) error {

	// KeyPair Class Config
	keypairCfg, err := config.LoadKeyPairClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found KeyPair class configuration for [" + class + "]!")

	if !dryRun {

		currentUser, _ := user.Current()
		sshLocation := currentUser.HomeDir + "/.ssh/"

		privateKeyPath := sshLocation + class + ".pem"
		publicKeyPath := sshLocation + class + ".pub"

		// Private Key
		privateKey := []byte(keypairCfg.PrivateKey1 + keypairCfg.PrivateKey2 + keypairCfg.PrivateKey3 + keypairCfg.PrivateKey4)

		if _, err := os.Stat(privateKeyPath); !os.IsNotExist(err) {
			terminal.ErrorLine("Local private key named [" + class + "] already exists!")

		} else if len(privateKey) < 1 {
			terminal.ErrorLine("Private key length is 0, not writing file!")

		} else {
			err = ioutil.WriteFile(privateKeyPath, privateKey, 0600)
			if err != nil {
				return err
			}

			terminal.Delta("Created private key at [" + privateKeyPath + "]")
		}

		// Public Key
		if _, err := os.Stat(publicKeyPath); !os.IsNotExist(err) {
			terminal.ErrorLine("Local public key named [" + class + "] already exists!")

		} else if len(keypairCfg.PublicKey) < 1 {
			terminal.ErrorLine("Public key length is 0, not writing file!")

		} else {
			err = ioutil.WriteFile(publicKeyPath, []byte(keypairCfg.PublicKey), 0600)
			if err != nil {
				return err
			}

			terminal.Delta("Created public key at [" + publicKeyPath + "]")
		}

	}

	return nil
}
