package aws

import (
	"errors"
	"os"
	"os/user"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/murdinc/terminal"
	"gopkg.in/ini.v1"
)

type awsmCreds struct {
	Profiles []Profile
}

// Profile represents a single AWS profile (an access key id and secret access key)
type Profile struct {
	Name            string `ini:"-"` // considered Sections in config file
	AccessKeyID     string `ini:"aws_access_key_id"`
	SecretAccessKey string `ini:"aws_secret_access_key"`
}

// CheckCreds Runs before everything, verifying we have proper authentication or asking us to set some up
func CheckCreds() (bool, string) {
	id, err := testCreds()

	if err != nil || len(id) == 0 {
		create := false

		// Try to read the config file
		cfg, err := readCreds()
		if err != nil || len(cfg.Profiles) == 0 {
			create = terminal.BoxPromptBool("The AWS Credentials config file is empty or missing!", "Do you want to add one now?")
		} else {
			create = terminal.BoxPromptBool("The AWS Credentials in your config file aren't working!", "Do you want to update it now?")
		}

		if !create {
			terminal.Information("Ok, maybe next time.. ")
			return false, ""
		}

		id := cfg.addCredsDialog()

		if len(id) > 0 {
			terminal.Notice("Success! Authenticated to AWS account: " + id)
		}

		return CheckCreds()
	}

	return true, id
}

// addCredsDialog is the dialog for the new creds setup
func (a *awsmCreds) addCredsDialog() string {
	// TODO prompt for default, or named alternatives

	accessKey := terminal.PromptString("What is your AWS Access Key Id?")
	secretKey := terminal.PromptString("What is your AWS Secret Access Key?")

	// Add Credentials to the ~/.aws/credentials file
	profile := Profile{Name: "default", AccessKeyID: accessKey, SecretAccessKey: secretKey}
	a.Profiles = append(a.Profiles, profile)

	err := a.SaveCreds()
	if err != nil {
		terminal.ErrorLine("There was a problem saving the config to [~/.aws/credentials]!")
		return ""
	}

	terminal.Delta("Checking...")

	id, err := testCreds()
	if err != nil {
		terminal.ErrorLine("There was a problem with your aws key and secret, please try again.")
		a.addCredsDialog()
	}
	return id
}

// testCreds verifies our credentials work and returns the current account ID
func testCreds() (string, error) {
	// Try to get the account id from our current users IAM creds
	iamUser, err := GetIAMUser("")
	if err == nil {
		// Parse the ARN to get the account ID
		parsedArn, err := ParseArn(iamUser.Arn)
		if err != nil || len(parsedArn.AccountID) == 0 {
			return "", err
		}

		return parsedArn.AccountID, nil
	}

	// Try to get the account if from the ec2metadata
	sess := session.Must(session.NewSession())
	svc := ec2metadata.New(sess)

	instanceDocument, err := svc.GetInstanceIdentityDocument()
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return "", errors.New(awsErr.Message())
		}
		return "", err
	}

	return instanceDocument.AccountID, nil
}

// readCreds reads in the config and returns a awsmCreds struct
func readCreds() (*awsmCreds, error) {
	// Reads in our config file
	config := new(awsmCreds)

	sep := string(os.PathSeparator)

	currentUser, _ := user.Current()
	configLocation := currentUser.HomeDir + sep + ".aws" + sep + "credentials"

	cfg, err := ini.Load(configLocation)
	if err != nil {
		return config, err
	}

	remotes := cfg.Sections()

	for _, remote := range remotes {

		// We dont want the default right now
		if remote.Name() == "DEFAULT" {
			continue
		}

		profile := new(Profile)

		err := remote.MapTo(profile)
		if err != nil {
			return config, err
		}

		profile.Name = remote.Name()
		config.Profiles = append(config.Profiles, *profile)
	}

	return config, err
}

// SaveCreds Saves our list of profiles into the config file
func (a *awsmCreds) SaveCreds() error {
	// Saves our config file
	currentUser, _ := user.Current()

	sep := string(os.PathSeparator)

	configFolder := currentUser.HomeDir + sep + ".aws" + sep
	configLocation := configFolder + "credentials"

	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		os.Mkdir(configFolder, os.FileMode(0755))
	}

	cfg := ini.Empty()

	for _, profile := range a.Profiles {
		err := cfg.Section(profile.Name).ReflectFrom(&profile)
		if err != nil {
			return err
		}
	}

	err := cfg.SaveToIndent(configLocation, "\t")
	if err != nil {
		return err
	}

	return nil
}

// TODO: Delete from config?
