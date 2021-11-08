package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lacework/go-sdk/generate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// iac-generate command is used to create IaC code for various environments
	generateTfCommand = &cobra.Command{
		Use:     "generate",
		Aliases: []string{"iac-generate", "iac"},
		Short:   "create iac code",
		Long:    "Create IaC content for various different cloud environments and configurations",
	}

	// aws command is used to generate TF code for aws
	generateAwsTfCommand = &cobra.Command{
		Use:   "aws",
		Short: "generate code for aws environment",
		Long:  "Genereate Terraform code for deploying into a new AWS enviornment.",
		RunE: func(_ *cobra.Command, args []string) error {
			if !cli.InteractiveMode() {
				return errors.New("interactive mode is disabled")
			}

			location, err := promptAwsGenerate()
			if err != nil {
				return errors.Wrap(err, "unable to create iac code")
			}

			cli.OutputHuman(fmt.Sprintf("Terraform Code generated at %s!\n", location))
			return nil
		},
	}
)

func init() {
	// add the iac-generate command
	rootCmd.AddCommand(generateTfCommand)

	// add sub-commands to the iac-generate command
	generateTfCommand.AddCommand(generateAwsTfCommand)
}

func promptAwsGenerate() (string, error) {
	questions := []*survey.Question{
		{
			Name:     "configureCloudtrail",
			Prompt:   &survey.Confirm{Message: "Enable Cloudtrail Integration?"},
			Validate: survey.Required,
		},
		{
			Name:     "configureConfig",
			Prompt:   &survey.Confirm{Message: "Enable Config Integration?"},
			Validate: survey.Required,
		},
	}

	answers := struct {
		ConfigureCloudtrail bool `survey:"configureCloudtrail"`
		ConfigureConfig     bool `survey:"configureConfig"`
	}{}

	ctQuestions := []*survey.Question{
		{
			Name:   "useConsolidatedCloudtrail",
			Prompt: &survey.Confirm{Message: "Use consolidated Cloudtrail?"},
		},
		{
			Name:   "awsRegion",
			Prompt: &survey.Input{Message: "(Optional) Specify the AWS region Cloudtrail, SNS, and S3 resources should use:"},
		},
		{
			Name:   "existingBucketArn",
			Prompt: &survey.Input{Message: "(Optional) Specify an existing bucket ARN used for Cloudtrail logs:"},
		},
		{
			Name:   "useExistingIamRole",
			Prompt: &survey.Confirm{Message: "Use an existing IAM Role?"},
		},
	}

	ctAnswers := struct {
		AwsRegion                 string `survey:"awsRegion"`
		ExistingBucketArn         string `survey:"existingBucketArn"`
		ExistingSnsTopicName      string `survey:"existingSnsTopicName"`
		UseConsolidatedCloudtrail bool   `survey:"useConsolidatedCloudtrail"`
		UseExistingIamRole        bool   `survey:"useExistingIamRole"`
	}{}

	err := survey.Ask(questions, &answers,
		survey.WithIcons(promptIconsFunc),
	)
	if err != nil {
		return "", err
	}

	if answers.ConfigureCloudtrail {
		err := survey.Ask(ctQuestions, &ctAnswers,
			survey.WithIcons(promptIconsFunc),
		)
		if err != nil {
			return "", err
		}
	}

	ctExistingIamAnswers := struct {
		ExistingIamRoleName       string `survey:"existingIamRoleName"`
		ExistingIamRoleArn        string `survey:"existingIamRoleArn"`
		ExistingIamRoleExternalId string `survey:"existingIamRoleExternalId"`
	}{}

	ctExistingIamQuestions := []*survey.Question{
		{
			Name:     "existingIamRoleName",
			Prompt:   &survey.Input{Message: "Specify an existing IAM role name for Cloudtrail access"},
			Validate: survey.Required,
		},
		{
			Name:     "existingIamRoleArn",
			Prompt:   &survey.Input{Message: "Specify an existing IAM role ARN for Cloudtrail access"},
			Validate: survey.Required,
		},
		{
			Name:     "existingIamRoleExternalId",
			Prompt:   &survey.Input{Message: "Specify the external ID to be used with the existing IAM role"},
			Validate: survey.Required,
		},
	}

	// If an existing IAM role is to be used, we need to collect the details
	if ctAnswers.UseExistingIamRole {
		err := survey.Ask(ctExistingIamQuestions, &ctExistingIamAnswers,
			survey.WithIcons(promptIconsFunc),
		)
		if err != nil {
			return "", err
		}
	}

	// If a new bucket is to be created; should the force destroy bit be set?
	var forceDestroyS3Bucket bool
	if ctAnswers.ExistingBucketArn != "" {
		err := survey.AskOne(
			&survey.Confirm{Message: "Should the new S3 bucket have force destroy enabled?"},
			&forceDestroyS3Bucket)
		if err != nil {
			return "", err
		}
	}

	// Let's collect up the other AWS accounts they would like to support
	collectMoreAccounts := false
	if ctAnswers.UseConsolidatedCloudtrail { // TODO This isn't the only time there might be sub-accounts
		err := survey.AskOne(
			&survey.Confirm{Message: "Are there additional AWS accounts to intergrate for Configuration?"},
			&collectMoreAccounts)
		if err != nil {
			return "", err
		}
	}

	type accountAnswers struct {
		AccountProfileName   string `survey:"accountProfileName"`
		AccountProfileRegion string `survey:"accountProfileRegion"`
	}
	accountQuestions := []*survey.Question{
		{
			Name:     "accountProfileName",
			Prompt:   &survey.Input{Message: "Supply the profile name for the AWS account"},
			Validate: survey.Required,
		},
		{
			Name:     "accountProfileRegion",
			Prompt:   &survey.Input{Message: "What region should be used for this account?"},
			Validate: survey.Required,
		},
	}

	// For each added account, collect it's profile name and the region that should be used
	accountDetails := map[string]string{}
	var mainAccountProfile string
	askAgain := true
	if collectMoreAccounts {
		// Determine the profile for the main account
		err := survey.AskOne(
			&survey.Input{Message: "What is the AWS profile name for the main account?"},
			&mainAccountProfile,
			survey.WithValidator(survey.Required),
		)
		if err != nil {
			return "", err
		}

		answers := accountAnswers{}
		for askAgain {
			err := survey.Ask(accountQuestions, &answers)
			if err != nil {
				return "", err
			}
			accountDetails[answers.AccountProfileName] = answers.AccountProfileRegion

			err = survey.AskOne(
				&survey.Confirm{Message: "Add another AWS account?"},
				&askAgain,
			)
			if err != nil {
				return "", err
			}
		}
	}

	// Generate TF Code
	cli.StartProgress("Generating Terraform Code...")
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail:       answers.ConfigureCloudtrail,
		ConfigureConfig:           answers.ConfigureConfig,
		AwsRegion:                 ctAnswers.AwsRegion,
		AwsProfile:                mainAccountProfile,
		ExistingIamRoleArn:        ctExistingIamAnswers.ExistingIamRoleArn,
		ExistingIamRoleName:       ctExistingIamAnswers.ExistingIamRoleName,
		ExistingIamRoleExternalId: ctExistingIamAnswers.ExistingIamRoleExternalId,
		UseExistingIamRole:        ctAnswers.UseExistingIamRole,
		ExistingBucketArn:         ctAnswers.ExistingBucketArn,
		ExistingSnsTopicArn:       ctAnswers.ExistingSnsTopicName,
		UseConsolidatedCloudtrail: ctAnswers.UseConsolidatedCloudtrail,
		ForceDestroyS3Bucket:      forceDestroyS3Bucket,
		ConfigureMoreAccounts:     collectMoreAccounts,
		Profiles:                  accountDetails,
	})

	// TODO Improve all this && Make output dir configurable
	// Write out
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	directory := filepath.FromSlash(fmt.Sprintf("%s/%s", dirname, "lacework"))
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.Mkdir(directory, 0700)
		if err != nil {
			return "", err
		}
	}

	location := fmt.Sprintf("%s/%s/main.tf", dirname, "lacework")
	err = os.WriteFile(
		filepath.FromSlash(location),
		[]byte(hcl),
		0700,
	)
	if err != nil {
		return "", err
	}

	cli.StopProgress()
	return location, err
}
