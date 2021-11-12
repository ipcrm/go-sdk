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
		Use:     "iac-generate",
		Aliases: []string{"iac-generate", "iac"},
		Short:   "create iac code",
		Long:    "Create IaC content for various different cloud environments and configurations",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			dirname, err := cmd.Flags().GetString("output")
			if err == nil {
				_, err := os.Stat(dirname)
				if err != nil {
					errors.Wrap(err, "could not access specified output location!")
				}
			}

			return nil
		},
	}

	// aws command is used to generate TF code for aws
	generateAwsTfCommand = &cobra.Command{
		Use:   "aws",
		Short: "generate code for aws environment",
		Long:  "Genereate Terraform code for deploying into a new AWS enviornment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate TF Code
			cli.StartProgress("Generating Terraform Code...")
			hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
				ConfigureCloudtrail:       generate.GenerateAwsCommandState.ConfigureCloudtrail,
				ConfigureConfig:           generate.GenerateAwsCommandState.ConfigureConfig,
				AwsRegion:                 generate.GenerateAwsCommandState.AwsRegion,
				AwsProfile:                generate.GenerateAwsCommandState.AwsProfile,
				UseExistingCloudtrail:     generate.GenerateAwsCommandState.UseExistingCloudtrail,
				ExistingBucketArn:         generate.GenerateAwsCommandState.ExistingBucketArn,
				ExistingIamRoleName:       generate.GenerateAwsCommandState.ExistingIamRoleName,
				ExistingIamRoleArn:        generate.GenerateAwsCommandState.ExistingIamRoleArn,
				ExistingIamRoleExternalId: generate.GenerateAwsCommandState.ExistingIamRoleExternalId,
				ExistingSnsTopicArn:       generate.GenerateAwsCommandState.ExistingSnsTopicArn,
				UseExistingIamRole:        generate.GenerateAwsCommandState.UseExistingIamRole,
				UseConsolidatedCloudtrail: generate.GenerateAwsCommandState.UseConsolidatedCloudtrail,
				ForceDestroyS3Bucket:      generate.GenerateAwsCommandState.ForceDestroyS3Bucket,
				Profiles:                  generate.GenerateAwsCommandState.Profiles,
				ConfigureMoreAccounts:     generate.GenerateAwsCommandState.ConfigureMoreAccounts,
				LaceworkProfile:           generate.GenerateAwsCommandState.LaceworkProfile,
			})

			return writeHclOutput(hcl, cmd)
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return promptAwsGenerate(generate.GenerateAwsCommandState)
		},
	}
)

func init() {
	// add the iac-generate command
	rootCmd.AddCommand(generateTfCommand)

	// Add global flags for iac generation
	generateTfCommand.PersistentFlags().String("output", "", "Location to write generated content")

	// add flags to sub commands
	// TODO Share the help with the interactive generation
	generateAwsTfCommand.PersistentFlags().BoolVar(
		&generate.GenerateAwsCommandState.ConfigureCloudtrailCli, "cloudtrail", false, "Configure Cloudtrail?")
	generateAwsTfCommand.PersistentFlags().BoolVar(
		&generate.GenerateAwsCommandState.ConfigureConfigCli, "config", false, "Enable Config Integration?")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.AwsRegion, "awsregion", "", "Specify AWS Region")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.AwsProfile, "awsprofile", "", "Specify AWS Profile")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.ExistingBucketArn,
		"existingbucketarn",
		"",
		"Specify existing Cloudtrail S3 bucket ARN")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.ExistingIamRoleArn,
		"existingiamrolearn",
		"",
		"Specify existing IAM role arn to use")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.ExistingIamRoleName,
		"existingiamrolename",
		"",
		"Specify existing IAM role name to use")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.ExistingIamRoleExternalId,
		"existingiamroleexternalid",
		"",
		"Specify existing IAM role external_id to use")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.ExistingIamRoleExternalId,
		"existingsnstopicarn",
		"",
		"Specify existing SNS topic ARN")
	generateAwsTfCommand.PersistentFlags().BoolVar(
		&generate.GenerateAwsCommandState.ConsolidatedCtCli,
		"consolidatedcloudtrail",
		false,
		"Use consolidated trail?")
	generateAwsTfCommand.PersistentFlags().BoolVar(
		&generate.GenerateAwsCommandState.ForceDestroyS3BucketCli,
		"forcedestroys3",
		false,
		"Enable force destroy S3 bucket?")
	generateAwsTfCommand.PersistentFlags().StringVar(
		&generate.GenerateAwsCommandState.ExistingIamRoleExternalId,
		"laceworkprofile",
		"",
		"Set the Lacework profile to use")

	// add sub-commands to the iac-generate command
	generateTfCommand.AddCommand(generateAwsTfCommand)
}

// Only prompt for an input if the CLI is interactive
func SurveyQuestionInteractiveOnly(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	if !cli.nonInteractive {
		err := survey.AskOne(p, response, opts...)
		if err != nil {
			return err
		}
	}
	return nil
}

// Only prompt for an input if the CLI is interactive and check is true
func SurveyQuestionWithValidation(check bool, p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	if check {
		return SurveyQuestionInteractiveOnly(p, response, opts...)
	}
	return nil
}

type SurveyQuestionWithValidationArgs struct {
	Prompt   survey.Prompt
	Check    bool
	Response interface{}
	Opts     []survey.AskOpt
}

// Prompt for many values at once
//
// checks: If supplied check(s) are true, questions will be asked
func SurveyMultipleQuestionWithValidation(questions []SurveyQuestionWithValidationArgs, checks ...bool) error {
	// Do validations pass?
	ok := true
	for _, v := range checks {
		if !v {
			ok = false
		}
	}

	// Ask questions
	if ok {
		for _, qs := range questions {
			if err := SurveyQuestionWithValidation(qs.Check, qs.Prompt, qs.Response, qs.Opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

// Write HCL output
func writeHclOutput(hcl string, cmd *cobra.Command) error {
	// Write out
	var dirname string
	dirname, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	if dirname == "" {
		dirname, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}

	directory := filepath.FromSlash(fmt.Sprintf("%s/%s", dirname, "lacework"))
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.Mkdir(directory, 0700)
		if err != nil {
			return err
		}
	}

	location := fmt.Sprintf("%s/%s/main.tf", dirname, "lacework")
	err = os.WriteFile(
		filepath.FromSlash(location),
		[]byte(hcl),
		0700,
	)
	if err != nil {
		return err
	}

	cli.StopProgress()
	cli.OutputHuman(fmt.Sprintf("Terraform Code generated at %s!\n", location))
	return nil
}
