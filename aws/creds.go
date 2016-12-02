package aws

import (
	"os/user"

	"github.com/aws/aws-sdk-go/aws/credentials"
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
func CheckCreds() bool {
	creds, err := testCreds()
	if err != nil || len(creds.ProviderName) == 0 {
		// Try to read the config file
		cfg, err := readCreds()
		if err != nil || len(cfg.Profiles) == 0 {

			// No Config Found, ask if we want to create one
			create := terminal.BoxPromptBool("No AWS Credentials found!", "Do you want to add them now?")
			if !create {
				terminal.Information("Ok then, maybe next time.. ")
				return false
			}
			cfg.addCredsDialog()
		}
	}
	return true
}

// addCredsDialog is the dialog for the new creds setup
func (a *awsmCreds) addCredsDialog() {

	// TODO prompt for default, or named alternatives

	accessKey := terminal.PromptString("What is your AWS Access Key Id?")
	secretKey := terminal.PromptString("What is your AWS Secret Access Key?")

	// Add Credentials to the ~/.aws/credentials file
	profile := Profile{Name: "default", AccessKeyID: accessKey, SecretAccessKey: secretKey}
	a.Profiles = append(a.Profiles, profile)

	err := a.SaveCreds()
	if err != nil {
		terminal.ErrorLine("There was a problem saving the config to [~/.aws/credentials]!")
	}

	creds, err := testCreds()
	if err != nil || len(creds.ProviderName) == 0 {
		terminal.ErrorLine("There was a problem with auth, please try again.")
		a.addCredsDialog()
	}
}

// testCreds verifies our credentials work
func testCreds() (credentials.Value, error) {
	// TODO more substiantial testing
	sess := session.New()
	return sess.Config.Credentials.Get()
}

// readCreds reads in the config and returns a awsmCreds struct
func readCreds() (*awsmCreds, error) {
	// Reads in our config file
	config := new(awsmCreds)

	currentUser, _ := user.Current()
	configLocation := currentUser.HomeDir + "/.aws/credentials"

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
	configLocation := currentUser.HomeDir + "/.aws/credentials"

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
