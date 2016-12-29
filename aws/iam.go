package aws

import (
	"os"
	"reflect"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/dustin/go-humanize"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// IAMUsers represents a slice of AWS IAM Users
type IAMUsers []IAMUser

// IAMUser represents a single IAM User
type IAMUser models.IAMUser

// IAMRoles represents a slice of AWS IAM Roles
type IAMRoles []IAMRole

// IAMRole represents a single IAM Role
type IAMRole models.IAMRole

// GetIAMUser returns a single IAM User that matches the provided username
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

// GetIAMUsers returns a list of IAM Users that match the provided search term
func GetIAMUsers(search string) (iamList *IAMUsers, err error) {
	svc := iam.New(session.New())
	result, err := svc.ListUsers(&iam.ListUsersInput{}) // TODO truncated?

	if err != nil {
		terminal.ShowErrorMessage("Error gathering IAM Users list", err.Error())
		return &IAMUsers{}, err
	}

	iam := make(IAMUsers, len(result.Users))
	for i, user := range result.Users {
		iam[i].Marshal(user)
	}

	iamList = new(IAMUsers)

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, ia := range iam {
			rIam := reflect.ValueOf(ia)

			for k := 0; k < rIam.NumField(); k++ {
				sVal := rIam.Field(k).String()

				if term.MatchString(sVal) {
					*iamList = append(*iamList, iam[i])
					continue Loop
				}
			}
		}
	} else {
		*iamList = append(*iamList, iam[:]...)
	}

	return iamList, nil
}

// GetIAMRole returns a single IAM Role that matches the provided name
func GetIAMRole(name string) (IAMRole, error) {
	svc := iam.New(session.New())

	params := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	resp, err := svc.GetRole(params)
	if err != nil {
		return IAMRole{}, err
	}

	role := new(IAMRole)
	role.Marshal(resp.Role)

	return *role, nil
}

// GetIAMRoles returns a list of IAM Roles that matches the provided name
func GetIAMRoles(search string) (iamRoleList *IAMRoles, err error) {
	svc := iam.New(session.New())
	result, err := svc.ListRoles(&iam.ListRolesInput{})

	if err != nil {
		terminal.ShowErrorMessage("Error gathering IAM Roles list", err.Error())
		return &IAMRoles{}, err
	}

	iam := make(IAMRoles, len(result.Roles))
	for i, role := range result.Roles {
		iam[i].Marshal(role)
	}

	iamRoleList = new(IAMRoles)

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, ia := range iam {
			rIam := reflect.ValueOf(ia)

			for k := 0; k < rIam.NumField(); k++ {
				sVal := rIam.Field(k).String()

				if term.MatchString(sVal) {
					*iamRoleList = append(*iamRoleList, iam[i])
					continue Loop
				}
			}
		}
	} else {
		*iamRoleList = append(*iamRoleList, iam[:]...)
	}

	return iamRoleList, nil
}

// Marshal parses the response from the aws sdk into an awsm IAM User
func (i *IAMUser) Marshal(user *iam.User) {
	i.UserName = aws.StringValue(user.UserName)
	i.UserID = aws.StringValue(user.UserId)
	i.CreateDate = aws.TimeValue(user.CreateDate) // robots
	i.CreatedHuman = humanize.Time(i.CreateDate)  // humans
	i.Arn = aws.StringValue(user.Arn)
	i.PasswordLastUsed = aws.TimeValue(user.PasswordLastUsed)   // robots
	i.PasswordLastUsedHuman = humanize.Time(i.PasswordLastUsed) // humans
}

// Marshal parses the response from the aws sdk into an awsm IAM Role
func (i *IAMRole) Marshal(user *iam.Role) {
	i.RoleName = aws.StringValue(user.RoleName)
	i.RoleID = aws.StringValue(user.RoleId)
	i.CreateDate = aws.TimeValue(user.CreateDate) // robots
	i.CreatedHuman = humanize.Time(i.CreateDate)  // humans
	i.Arn = aws.StringValue(user.Arn)
}

// PrintTable Prints an ascii table of the list of IAM Users
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

// PrintTable Prints an ascii table of the list of IAM Roles
func (i *IAMRoles) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No IAM Roles Found!")
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

// CreateIAMUser creates a new IAM User with the provided username and path
func CreateIAMUser(username, path string) error {

	svc := iam.New(session.New())

	params := &iam.CreateUserInput{
		UserName: aws.String(username),
	}

	if path != "" {
		params.Path = aws.String(path)
	}
	_, err := svc.CreateUser(params)
	if err == nil {
		terminal.Information("Done!")
	}

	return err
}

// DeleteIAMUsers deletes one or more IAM Users that match the provided username
func DeleteIAMUsers(username string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	userList, err := GetIAMUsers(username)
	if err != nil {
		terminal.ErrorLine("Error gathering IAM Role list")
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

	if !dryRun {
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
			terminal.Delta("Deleted IAM User [" + user.UserName + "]!")
		}

		terminal.Information("Done!")

	}

	return
}

// DeleteIAMRoles deletes one or more IAM Roles that match the provided name
func DeleteIAMRoles(name string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	roleList, err := GetIAMRoles(name)
	if err != nil {
		terminal.ErrorLine("Error gathering IAM Role list")
		return
	}

	if len(*roleList) > 0 {
		// Print the table
		roleList.PrintTable()
	} else {
		terminal.ErrorLine("No IAM Users found, Aborting!")
		return
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these IAM Users?") {
		terminal.ErrorLine("Aborting!")
		return
	}

	if !dryRun {
		// Delete 'Em
		for _, role := range *roleList {
			svc := iam.New(session.New())

			params := &iam.DeleteRoleInput{
				RoleName: aws.String(role.RoleName),
			}
			_, err := svc.DeleteRole(params)
			if err != nil {
				terminal.ErrorLine("Error while deleting IAM Role [" + role.RoleName + "], Aborting!")
				return err
			}
			terminal.Delta("Deleted IAM Role [" + role.RoleName + "]!")
		}

		terminal.Information("Done!")
	}

	return
}
