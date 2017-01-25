package aws

import (
	"errors"
	"net/url"
	"os"
	"reflect"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
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

// IAMPolicies represents a slice of AWS IAM Policy
type IAMPolicies []IAMPolicy

// IAMRole represents a single IAM Role
type IAMRole models.IAMRole

// IAMPolicy represents a single IAM Policy
type IAMPolicy models.IAMPolicy

// IAMPolicyDocument represents a single IAM Policy
type IAMPolicyDocument models.IAMPolicyDocument

// IAMProfile represents a slice of AWS IAM Profiles
type IAMInstanceProfiles []IAMInstanceProfile

// IAMRole represents a single IAM Profile
type IAMInstanceProfile models.IAMInstanceProfile

// GetIAMUser returns a single IAM User that matches the provided username
func GetIAMUser(username string) (IAMUser, error) {
	svc := iam.New(session.New())

	params := &iam.GetUserInput{}

	if username != "" {
		params.UserName = aws.String(username)
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

// GetIAMRolePolicyNames returns the names of IAM Role Policies that are embedded in the provided IAM Role
func GetIAMRolePolicyNames(roleName string) ([]string, error) {
	svc := iam.New(session.New())

	params := &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	resp, err := svc.ListRolePolicies(params)
	if err != nil {
		return []string{}, err
	}

	return aws.StringValueSlice(resp.PolicyNames), nil
}

// GetIAMAttachedRolePolicyNames returns the names of IAM Role Policies that are attached to the provided IAM Role
func GetIAMAttachedRolePolicyARNs(roleName string) ([]string, error) {
	svc := iam.New(session.New())

	params := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	resp, err := svc.ListAttachedRolePolicies(params)
	if err != nil {
		return []string{}, err
	}

	policyArns := make([]string, len(resp.AttachedPolicies))

	for i, policy := range resp.AttachedPolicies {
		policyArns[i] = aws.StringValue(policy.PolicyArn)
	}

	return policyArns, nil
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

// GetIAMRoleByName returns an IAM Role that matches the provided name
func GetIAMRoleByName(roleName string) (iamRole IAMRole, err error) {

	iamRoles, err := GetIAMRoles("")
	if err != nil {
		return iamRole, nil
	}

	for _, role := range *iamRoles {
		if role.RoleName == roleName {
			return role, nil
		}
	}

	return iamRole, errors.New("No IAM Role found matching [" + roleName + "]!")
}

// GetIAMPolicyByName returns an IAM Policy that matches the provided name
func GetIAMPolicyByName(policyName string) (iamPolicy IAMPolicy, err error) {

	iamPolicies, err := GetIAMPolicies("")
	if err != nil {
		return iamPolicy, nil
	}

	for _, policy := range *iamPolicies {
		if policy.PolicyName == policyName {
			return policy, nil
		}
	}

	return iamPolicy, errors.New("No IAM Policy found matching [" + policyName + "]!")
}

// GetIAMInstanceProfileByName returns an IAM Instance Profile that matches the provided name
func GetIAMInstanceProfileByName(instanceProfileName string) (iamInstanceProfile IAMInstanceProfile, err error) {

	iamInstanceProfiles, err := GetIAMInstanceProfiles("")
	if err != nil {
		return iamInstanceProfile, nil
	}

	for _, profile := range *iamInstanceProfiles {
		if profile.ProfileName == instanceProfileName {
			return profile, nil
		}
	}

	return iamInstanceProfile, errors.New("No IAM Instance Profile found matching [" + instanceProfileName + "]!")
}

// GetIAMPolicy returns a single IAM Policy given a provided search term
func GetIAMPolicyDocument(search string, version string) (IAMPolicyDocument, error) {

	policies, err := GetIAMPolicies(search)
	if err != nil {
		return IAMPolicyDocument{}, err
	}

	policyCount := len(*policies)

	if policyCount == 0 {
		return IAMPolicyDocument{}, errors.New("No IAM Policies found for the search term provided")
	} else if policyCount > 1 {
		policies.PrintTable()
		return IAMPolicyDocument{}, errors.New("Please limit your search term to just one Policy")
	}

	policy := (*policies)[0]

	if version == "" {
		version = policy.DefaultVersionId
	}

	svc := iam.New(session.New())
	params := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(policy.Arn),
		VersionId: aws.String(version),
	}

	resp, err := svc.GetPolicyVersion(params)
	if err != nil {
		return IAMPolicyDocument{}, err
	}

	policyDocument := new(IAMPolicyDocument)
	policyDocument.Marshal(resp.PolicyVersion)

	return *policyDocument, nil
}

// GetIAMPolicyByARN returns a single IAM Policy that matches the provided ARN
func GetIAMPolicyByARN(policyARN string) (iamPolicy *IAMPolicy, err error) {
	svc := iam.New(session.New())

	params := &iam.GetPolicyInput{
		PolicyArn: aws.String(policyARN),
	}

	result, err := svc.GetPolicy(params)

	if err != nil {
		terminal.ShowErrorMessage("Error fetching IAM Policy", err.Error())
		return &IAMPolicy{}, err
	}

	iamPolicy = new(IAMPolicy)
	iamPolicy.Marshal(result.Policy)

	return iamPolicy, nil
}

// GetIAMPolicies returns a list of IAM Policies that matches the provided name
func GetIAMPolicies(search string) (iamPolicyList *IAMPolicies, err error) {
	svc := iam.New(session.New())
	result, err := svc.ListPolicies(&iam.ListPoliciesInput{})

	if err != nil {
		terminal.ShowErrorMessage("Error gathering IAM Policy list", err.Error())
		return &IAMPolicies{}, err
	}

	iam := make(IAMPolicies, len(result.Policies))
	for i, policy := range result.Policies {
		iam[i].Marshal(policy)
	}

	iamPolicyList = new(IAMPolicies)

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, ia := range iam {
			rIam := reflect.ValueOf(ia)

			for k := 0; k < rIam.NumField(); k++ {
				sVal := rIam.Field(k).String()

				if term.MatchString(sVal) {
					*iamPolicyList = append(*iamPolicyList, iam[i])
					continue Loop
				}
			}
		}
	} else {
		*iamPolicyList = append(*iamPolicyList, iam[:]...)
	}

	return iamPolicyList, nil
}

// GetIAMProfile returns a single IAM Profile that matches the provided name
func GetIAMInstanceProfile(name string) (IAMInstanceProfile, error) {
	svc := iam.New(session.New())

	params := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}

	resp, err := svc.GetInstanceProfile(params)
	if err != nil {
		return IAMInstanceProfile{}, err
	}

	profile := new(IAMInstanceProfile)
	profile.Marshal(resp.InstanceProfile)

	return *profile, nil
}

// GetIAMInstanceProfiles returns a list of IAM Profiles that matches the provided name
func GetIAMInstanceProfiles(search string) (iamProfileList *IAMInstanceProfiles, err error) {
	svc := iam.New(session.New())
	result, err := svc.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})

	if err != nil {
		terminal.ShowErrorMessage("Error gathering IAM Instance Profile list", err.Error())
		return &IAMInstanceProfiles{}, err
	}

	iam := make(IAMInstanceProfiles, len(result.InstanceProfiles))
	for i, profile := range result.InstanceProfiles {
		iam[i].Marshal(profile)
	}

	iamProfileList = new(IAMInstanceProfiles)

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, ia := range iam {
			rIam := reflect.ValueOf(ia)

			for k := 0; k < rIam.NumField(); k++ {
				sVal := rIam.Field(k).String()

				if term.MatchString(sVal) {
					*iamProfileList = append(*iamProfileList, iam[i])
					continue Loop
				}
			}
		}
	} else {
		*iamProfileList = append(*iamProfileList, iam[:]...)
	}

	return iamProfileList, nil
}

// GetIAMInstanceProfiles returns a list of IAM Profiles that matches the provided name
func GetIAMInstanceProfilesForRole(roleName string) (iamInstanceProfileList IAMInstanceProfiles, err error) {
	svc := iam.New(session.New())

	params := &iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	}

	result, err := svc.ListInstanceProfilesForRole(params)

	if err != nil {
		terminal.ShowErrorMessage("Error gathering IAM Instance Profiles list for Role ["+roleName+"]", err.Error())
		return IAMInstanceProfiles{}, err
	}

	iamInstanceProfileList = make(IAMInstanceProfiles, len(result.InstanceProfiles))
	for i, profile := range result.InstanceProfiles {
		iamInstanceProfileList[i].Marshal(profile)
	}

	return iamInstanceProfileList, nil
}

// RemoveIAMRoleFromInstanceProfile removes an IAM Role from an Instance Profile
func RemoveIAMRoleFromInstanceProfile(roleName, instanceProfileName string) error {
	svc := iam.New(session.New())

	params := &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
		RoleName:            aws.String(roleName),
	}

	_, err := svc.RemoveRoleFromInstanceProfile(params)
	if err != nil {
		terminal.ShowErrorMessage("Error removing IAM Role ["+roleName+"] from Instance Profile ["+instanceProfileName+"]", err.Error())
		return err
	}

	terminal.Delta("Removed IAM Role [" + roleName + "] from Instance Profile [" + instanceProfileName + "]!")

	return nil
}

// DetachIAMRolePolicy detaches an IAM Role from a policy
func DetachIAMRolePolicy(roleName, policyArn string) error {
	svc := iam.New(session.New())

	params := &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	}

	_, err := svc.DetachRolePolicy(params)
	if err != nil {
		terminal.ShowErrorMessage("Error detaching IAM Policy ["+policyArn+"] from Role ["+roleName+"]", err.Error())
		return err
	}

	terminal.Delta("Detached IAM Role [" + roleName + "] from Policy [" + policyArn + "]!")

	return nil
}

// AttachIAMRolePolicy attaches an IAM Role to a policy
func AttachIAMRolePolicy(roleName, policy string, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Check that we have a role with this name
	roleList, err := GetIAMRoles(roleName)
	if err != nil {
		return err
	}

	roleCount := len(*roleList)

	if roleCount < 1 {
		return errors.New("No IAM Roles found matching [" + roleName + "]")
	} else if roleCount > 1 {
		roleList.PrintTable()
		return errors.New("Please limit your search term to just one IAM Role")
	}

	policies, err := GetIAMPolicies(policy)
	if err != nil {
		return err
	}
	if len(*policies) < 1 {
		return errors.New("No IAM Policies found matching [" + policy + "]")
	}

	terminal.Notice("Role:")
	roleList.PrintTable()
	terminal.Notice("Policies:")
	policies.PrintTable()

	// Confirm
	if !terminal.PromptBool("Are you sure you want to these policies to this IAM Role?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	if !dryRun {
		// Attach 'Em
		for _, policy := range *policies {
			err := AttachIAMRolePolicyByARN(roleName, policy.Arn, dryRun)
			if err != nil {
				return err
			}
		}
	}

	terminal.Information("Done!")

	return nil
}

func AttachIAMRolePolicyByARN(roleName, policyARN string, dryRun bool) error {
	attachedPolicyARNs, err := GetIAMAttachedRolePolicyARNs(roleName)
	if err != nil {
		terminal.ShowErrorMessage("Error attaching IAM Policy ["+policyARN+"] to Role ["+roleName+"]", err.Error())
		return err
	}

	for _, attachedPolicyARN := range attachedPolicyARNs {
		if attachedPolicyARN == policyARN {
			terminal.Notice("IAM Policy [" + policyARN + "] is already attached to to Role [" + roleName + "], skipping")
			return nil
		}
	}

	if !dryRun {
		svc := iam.New(session.New())

		params := &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyARN),
		}

		_, err = svc.AttachRolePolicy(params)
		if err != nil {
			terminal.ShowErrorMessage("Error attaching IAM Policy ["+policyARN+"] to Role ["+roleName+"]", err.Error())
			return err
		}

		terminal.Delta("Attached IAM Policy [" + policyARN + "] to Role [" + roleName + "]!")
	}
	terminal.Information("Done!")

	return nil
}

func AddIAMRoleToInstanceProfile(roleName, instanceProfileName string, dryRun bool) error {

	instProfiles, err := GetIAMInstanceProfilesForRole(roleName)
	if err != nil {
		terminal.ShowErrorMessage("Error adding IAM Role ["+roleName+"] to Instance Profile ["+instanceProfileName+"]", err.Error())
		return err
	}
	for _, instProfile := range instProfiles {
		if instProfile.ProfileName == instanceProfileName {
			terminal.Notice("IAM Role [" + roleName + "] has already been added to to Instance Profile [" + instanceProfileName + "], skipping.")
			return nil
		}
	}

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	if !dryRun {
		svc := iam.New(session.New())

		params := &iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(instanceProfileName),
			RoleName:            aws.String(roleName),
		}

		_, err := svc.AddRoleToInstanceProfile(params)
		if err != nil {
			terminal.ShowErrorMessage("Error adding IAM Role ["+roleName+"] to Instance Profile ["+instanceProfileName+"]", err.Error())
			return err
		}

		terminal.Delta("Attached IAM Role [" + roleName + "] to Instance Profile [" + instanceProfileName + "]!")
	}
	return nil
}

// Marshal parses the response from the aws sdk into an awsm IAM User
func (i *IAMUser) Marshal(user *iam.User) {
	i.UserName = aws.StringValue(user.UserName)
	i.UserID = aws.StringValue(user.UserId)
	i.CreateDate = aws.TimeValue(user.CreateDate)
	i.Arn = aws.StringValue(user.Arn)
	i.PasswordLastUsed = aws.TimeValue(user.PasswordLastUsed)
}

// Marshal parses the response from the aws sdk into an awsm IAM Role
func (i *IAMRole) Marshal(user *iam.Role) {
	doc, _ := url.Parse(aws.StringValue(user.AssumeRolePolicyDocument))
	docStr := doc.Path

	i.RoleName = aws.StringValue(user.RoleName)
	i.RoleID = aws.StringValue(user.RoleId)
	i.CreateDate = aws.TimeValue(user.CreateDate)
	i.Arn = aws.StringValue(user.Arn)
	i.AssumeRolePolicyDocument = docStr
}

// Marshal parses the response from the aws sdk into an awsm IAM Policy
func (i *IAMPolicy) Marshal(policy *iam.Policy) {
	i.PolicyName = aws.StringValue(policy.PolicyName)
	i.PolicyID = aws.StringValue(policy.PolicyId)
	i.Description = aws.StringValue(policy.Description)
	i.IsAttachable = aws.BoolValue(policy.IsAttachable)
	i.CreateDate = aws.TimeValue(policy.CreateDate)
	i.AttachmentCount = int(aws.Int64Value(policy.AttachmentCount))
	i.DefaultVersionId = aws.StringValue(policy.DefaultVersionId)
	i.Arn = aws.StringValue(policy.Arn)
}

// Marshal parses the response from the aws sdk into an awsm IAM Policy Document
func (i *IAMPolicyDocument) Marshal(policyVersion *iam.PolicyVersion) {
	doc, _ := url.Parse(aws.StringValue(policyVersion.Document))
	docStr := doc.Path

	i.Document = docStr
	i.CreateDate = aws.TimeValue(policyVersion.CreateDate)
	i.IsDefaultVersion = aws.BoolValue(policyVersion.IsDefaultVersion)
	i.VersionId = aws.StringValue(policyVersion.VersionId)
}

// Marshal parses the response from the aws sdk into an awsm IAM Role
func (i *IAMInstanceProfile) Marshal(profile *iam.InstanceProfile) {
	i.ProfileName = aws.StringValue(profile.InstanceProfileName)
	i.ProfileID = aws.StringValue(profile.InstanceProfileId)
	i.CreateDate = aws.TimeValue(profile.CreateDate)
	i.Arn = aws.StringValue(profile.Arn)
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

// PrintTable Prints an ascii table of the list of IAM Policies
func (i *IAMPolicies) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No IAM Policies Found!")
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

// PrintTable Prints an ascii table of the list of IAM Instance Profiles
func (i *IAMInstanceProfiles) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No IAM Instance Profiles Found!")
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

// CreateIAMUser creates a new IAM User with the provided username and path
func CreateIAMPolicy(policyName, policyDocument, path, description string, dryRun bool) (string, error) {

	policy, _ := GetIAMPolicyByName(policyName)
	if policy.Arn != "" {
		terminal.Notice("IAM Policy named [" + policyName + "] already exists, skipping.")
		return policy.Arn, nil
	}

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	svc := iam.New(session.New())

	params := &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	}

	if path != "" {
		params.Path = aws.String(path)
	}
	if description != "" {
		params.Description = aws.String(description)
	}

	if !dryRun {
		resp, err := svc.CreatePolicy(params)
		if err != nil {
			return "", err
		}

		policyName := aws.StringValue(resp.Policy.PolicyName)
		policyArn := aws.StringValue(resp.Policy.Arn)

		terminal.Delta("Created IAM Policy named [" + policyName + "] with ARN [" + policyArn + "]")

		return policyArn, err
	}

	terminal.Information("Done!")

	return "", nil
}

// CreateIAMRole creates a new IAM Role with the provided name, policyDocument, and optional path
func CreateIAMRole(roleName, rolePolicyDocument, path string, dryRun bool) (string, error) {

	role, _ := GetIAMRoleByName(roleName)
	if role.Arn != "" {
		terminal.Notice("IAM Role named [" + roleName + "] already exists, skipping.")
		return role.Arn, nil
	}

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	svc := iam.New(session.New())

	params := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(rolePolicyDocument),
		RoleName:                 aws.String(roleName),
	}

	if path != "" {
		params.Path = aws.String(path)
	}

	if !dryRun {
		resp, err := svc.CreateRole(params)
		if err != nil {
			return "", err
		}

		roleName := aws.StringValue(resp.Role.RoleName)
		roleArn := aws.StringValue(resp.Role.Arn)

		terminal.Delta("Created IAM Role named [" + roleName + "] with ARN [" + roleArn + "]")

		return roleArn, err
	}

	terminal.Information("Done!")

	return "", nil
}

// CreateIAMInstanceProfile creates a new IAM Instance Profile with the provided name, and optional path
func CreateIAMInstanceProfile(instanceProfileName, path string, dryRun bool) (string, error) {

	instanceProfile, _ := GetIAMInstanceProfileByName(instanceProfileName)
	if instanceProfile.Arn != "" {
		terminal.Notice("IAM Instance Profile named [" + instanceProfileName + "] already exists, skipping.")
		return instanceProfile.Arn, nil
	}

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	svc := iam.New(session.New())

	params := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	}

	if path != "" {
		params.Path = aws.String(path)
	}

	if !dryRun {
		resp, err := svc.CreateInstanceProfile(params)
		if err != nil {
			return "", err
		}

		ipName := aws.StringValue(resp.InstanceProfile.InstanceProfileName)
		ipArn := aws.StringValue(resp.InstanceProfile.Arn)

		terminal.Delta("Created IAM Instance Profile named [" + ipName + "] with ARN [" + ipArn + "]")

		return ipArn, err
	}

	terminal.Information("Done!")

	return "", nil
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
	}

	terminal.Information("Done!")

	return
}

// DeleteIAMRoles deletes one or more IAM Roles that match the provided name
func DeleteIAMRoles(name string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	roleList, err := GetIAMRoles(name)
	if err != nil {
		terminal.ErrorLine("Error gathering IAM Role list")
		return err
	}

	if len(*roleList) > 0 {
		// Print the table
		roleList.PrintTable()
	} else {
		terminal.ErrorLine("No IAM Users found, Aborting!")
		return nil
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these IAM Roles?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	if !dryRun {
		// Delete 'Em
		for _, role := range *roleList {
			svc := iam.New(session.New())

			// Get the instance profiles for this role
			instProfiles, err := GetIAMInstanceProfilesForRole(role.RoleName)
			// Remove the role from these instance profiles
			for _, instProfile := range instProfiles {
				err = RemoveIAMRoleFromInstanceProfile(role.RoleName, instProfile.ProfileName)
				if err != nil {
					return err
				}
			}

			// Get the attached role Policies for this role
			rolePolicyARNS, err := GetIAMAttachedRolePolicyARNs(role.RoleName)
			// Detach the role from these role policies
			for _, rolePolicyARN := range rolePolicyARNS {
				err = DetachIAMRolePolicy(role.RoleName, rolePolicyARN)
				if err != nil {
					return err
				}
			}

			params := &iam.DeleteRoleInput{
				RoleName: aws.String(role.RoleName),
			}
			_, err = svc.DeleteRole(params)
			if err != nil {
				terminal.ErrorLine("Error while deleting IAM Role [" + role.RoleName + "], Aborting!")
				return err
			}
			terminal.Delta("Deleted IAM Role [" + role.RoleName + "]!")
		}
	}

	terminal.Information("Done!")

	return nil
}

// DeleteIAMInstanceProfiles deletes one or more IAM Instance Profiles that match the provided search term
func DeleteIAMInstanceProfiles(search string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	instProfileList, err := GetIAMInstanceProfiles(search)
	if err != nil {
		terminal.ErrorLine("Error gathering IAM Instance Profiles list")
		return err
	}

	if len(*instProfileList) > 0 {
		// Print the table
		instProfileList.PrintTable()
	} else {
		terminal.ErrorLine("No IAM Instance Profiles found, Aborting!")
		return nil
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these IAM Instance Profiles?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	if !dryRun {
		// Delete 'Em
		for _, instProfile := range *instProfileList {
			svc := iam.New(session.New())

			params := &iam.DeleteInstanceProfileInput{
				InstanceProfileName: aws.String(instProfile.ProfileName),
			}
			_, err = svc.DeleteInstanceProfile(params)
			if err != nil {
				terminal.ErrorLine("Error while deleting IAM Instance Profile [" + instProfile.ProfileName + "], Aborting!")
				return err
			}
			terminal.Delta("Deleted IAM Instance Profile [" + instProfile.ProfileName + "]!")
		}
	}

	terminal.Information("Done!")

	return nil
}

// DeleteIAMPolicies deletes one or more IAM Policies that match the provided name
func DeleteIAMPolicies(name string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	policyList, err := GetIAMPolicies(name)
	if err != nil {
		terminal.ErrorLine("Error gathering IAM Policy list")
		return
	}

	if len(*policyList) > 0 {
		// Print the table
		policyList.PrintTable()
	} else {
		terminal.ErrorLine("No IAM Policies found, Aborting!")
		return
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these IAM Policies?") {
		terminal.ErrorLine("Aborting!")
		return
	}

	if !dryRun {
		// Delete 'Em
		for _, policy := range *policyList {
			svc := iam.New(session.New())

			params := &iam.DeletePolicyInput{
				PolicyArn: aws.String(policy.Arn),
			}
			_, err := svc.DeletePolicy(params)
			if err != nil {
				terminal.ErrorLine("Error while deleting IAM Policy [" + policy.PolicyName + "], Aborting!")
				return err
			}
			terminal.Delta("Deleted IAM Policy [" + policy.PolicyName + "]!")
		}
	}

	terminal.Information("Done!")

	return
}
