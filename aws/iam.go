package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
)

type IAMUsers []IAM

type IAM struct {
	UserName         string
	UserId           string
	CreateDate       string
	Arn              string
	PasswordLastUsed string
}

func GetIAMUser(username string) (IAM, error) {
	svc := iam.New(session.New())

	params := &iam.GetUserInput{
		UserName: aws.String(username),
	}

	resp, err := svc.GetUser(params)
	if err != nil {
		return IAM{}, err
	}

	user := IAM{
		UserName:   aws.StringValue(resp.User.UserName),
		UserId:     aws.StringValue(resp.User.UserId),
		CreateDate: resp.User.CreateDate.String(),
		Arn:        aws.StringValue(resp.User.Arn),
		//PasswordLastUsed: user.PasswordLastUsed.String(), // TODO why dont some users have this, and why does it fail when its not present?
	}

	return user, nil
}

func GetIAMUsers(search string) (*IAMUsers, error) {
	iamList := new(IAMUsers)

	svc := iam.New(session.New())
	result, err := svc.ListUsers(&iam.ListUsersInput{}) // TODO truncated?

	if err != nil {
		terminal.ShowErrorMessage("Error gathering IAM Users list", err.Error())
		return iamList, err
	}

	iam := make(IAMUsers, len(result.Users))
	for i, user := range result.Users {
		iam[i] = IAM{
			UserName:   aws.StringValue(user.UserName),
			UserId:     aws.StringValue(user.UserId),
			CreateDate: user.CreateDate.String(),
			Arn:        aws.StringValue(user.Arn),
			//PasswordLastUsed: user.PasswordLastUsed.String(), // TODO why dont some users have this, and why does it fail when its not present?
		}
	}
	*iamList = append(*iamList, iam[:]...)

	return iamList, nil
}

func (i *IAMUsers) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.UserName,
			val.UserId,
			val.CreateDate,
			val.Arn,
			//val.PasswordLastUsed,
		}
	}

	table.SetHeader([]string{"User Name", "Id", "Creation Date", "Arn"}) //, "Password Last Used"})

	table.AppendBulk(rows)
	table.Render()
}

func CreateIAMUser(username, path string) error {

	svc := iam.New(session.New())

	params := &iam.CreateUserInput{
		UserName: aws.String(username),
		Path:     aws.String(path),
	}
	_, err := svc.CreateUser(params)
	if err == nil {
		terminal.Information("Done!")
	}

	return err
}

func DeleteIAMUser(username string) (err error) {

	userList, err := GetIAMUsers(username)
	if err != nil {
		terminal.ErrorLine("Error gathering SimpleDB domains list")
		return
	}

	if len(*userList) > 0 {
		// Print the table
		userList.PrintTable()
	} else {
		terminal.ErrorLine("No IAM Users found, Aborting!")
		return
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these IAM Users?") {
		terminal.ErrorLine("Aborting!")
		return
	}

	// Delete 'Em
	for _, user := range *userList {
		svc := iam.New(session.New())

		params := &iam.DeleteUserInput{
			UserName: aws.String(user.UserName),
		}
		_, err := svc.DeleteUser(params)
		if err != nil {
			terminal.ErrorLine("Error while deleting IAM User [" + user.UserName + "], Aborting!")
			return err
		}
		terminal.Information("Deleted IAM User [" + user.UserName + "]!")
	}

	terminal.Information("Done!")

	return
}
