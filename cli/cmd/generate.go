package cmd

import (
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

			response, err := promptAwsGenerate()
			if err != nil {
				return errors.Wrap(err, "unable to create iac code")
			}

			cli.OutputHuman(response)
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
			Name:   "awsRegion",
			Prompt: &survey.Input{Message: "Specify the AWS region Cloudtrail, SNS, and S3 resources should use"},
		},
		{
			Name:   "existingBucketArn",
			Prompt: &survey.Input{Message: "Specify an existing bucket ARN used for Cloudtrail logs"},
		},
		{
			Name:   "useConsolidatedCloudtrail",
			Prompt: &survey.Confirm{Message: "Use consolidated Cloudtrail?"},
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

	if ctAnswers.UseExistingIamRole {
		err := survey.Ask(ctExistingIamQuestions, &ctExistingIamAnswers,
			survey.WithIcons(promptIconsFunc),
		)
		if err != nil {
			return "", err
		}
	}

	var forceDestroyS3Bucket bool
	if ctAnswers.ExistingBucketArn != "" {
		survey.AskOne(&survey.Confirm{Message: "Should the new S3 bucket have force destroy enabled?"}, forceDestroyS3Bucket)
	}

	cli.StartProgress(" Generating Terraform Code...")
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail:       answers.ConfigureCloudtrail,
		ConfigureConfig:           answers.ConfigureConfig,
		AwsRegion:                 ctAnswers.AwsRegion,
		ExistingIamRoleArn:        ctExistingIamAnswers.ExistingIamRoleArn,
		ExistingIamRoleName:       ctExistingIamAnswers.ExistingIamRoleName,
		ExistingIamRoleExternalId: ctExistingIamAnswers.ExistingIamRoleExternalId,
		UseExistingIamRole:        ctAnswers.UseExistingIamRole,
		ExistingBucketArn:         ctAnswers.ExistingBucketArn,
		ExistingSnsTopicArn:       ctAnswers.ExistingSnsTopicName,
		UseConsolidatedCloudtrail: ctAnswers.UseConsolidatedCloudtrail,
		ForceDestroyS3Bucket:      forceDestroyS3Bucket,
	})

	cli.StopProgress()
	return hcl, err
}
