package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/dustin/go-humanize"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type IAMUsers []IAMUser

type IAMUser models.IAMUser

func GetIAMUser(username string) (IAMUser, error) {
	svc := iam.New(session.New())

	params := &iam.GetUserInput{
		UserName: aws.String(username),
	}

	resp, err := svc.GetUser(params)
	if err != nil {
		return IAMUser{}, err
	}

	user := new(IAMUser)
	user.Marshal(resp.User)

	return *user, nil
}

func GetIAMUsers(search string) (*IAMUsers, error) {
	svc := iam.New(session.New())
	result, err := svc.ListUsers(&iam.ListUsersInput{}) // TODO truncated?

	if err != nil {
		terminal.ShowErrorMessage("Error gathering IAM Users list", err.Error())
		return &IAMUsers{}, err
	}

	iamList := make(IAMUsers, len(result.Users))
	for i, user := range result.Users {
		iamList[i].Marshal(user)
	}

	return &iamList, nil
}

func (i *IAMUser) Marshal(user *iam.User) {
	i.UserName = aws.StringValue(user.UserName)
	i.UserId = aws.StringValue(user.UserId)
	i.CreateDate = aws.TimeValue(user.CreateDate) // robots
	i.CreatedHuman = humanize.Time(i.CreateDate)  // humans
	i.Arn = aws.StringValue(user.Arn)
	i.PasswordLastUsed = aws.TimeValue(user.PasswordLastUsed)   // robots
	i.PasswordLastUsedHuman = humanize.Time(i.PasswordLastUsed) // humans
}

func (i *IAMUsers) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No IAM Users Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, user := range *i {
		models.ExtractAwsmTable(index, user, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
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

func DeleteIAMUsers(username string) (err error) {

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
