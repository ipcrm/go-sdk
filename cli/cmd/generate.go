package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lacework/go-sdk/generate"
	"github.com/spf13/cobra"
)

var (
	// iac-generate command is used to create IaC code for various environments
	generateTfCommand = &cobra.Command{
		Use:     "iac-generate",
		Aliases: []string{"iac-generate", "iac"},
		Short:   "create iac code",
		Long:    "Create IaC content for various different cloud environments and configurations",
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
				UseConsolidatedCloudtrail: generate.GenerateAwsCommandState.UseConsolidatedCloudtrail,
				ForceDestroyS3Bucket:      generate.GenerateAwsCommandState.ForceDestroyS3Bucket,
				Profiles:                  generate.GenerateAwsCommandState.Profiles,
				ConfigureMoreAccounts:     generate.GenerateAwsCommandState.ConfigureMoreAccounts,
				LaceworkProfile:           generate.GenerateAwsCommandState.LaceworkProfile,
			})

			// TODO Improve all this && Make output dir configurable
			// Write out
			dirname, err := os.UserHomeDir()
			if err != nil {
				return err
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
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return promptAwsGenerate(generate.GenerateAwsCommandState)
		},
	}
)

func init() {
	// add the iac-generate command
	rootCmd.AddCommand(generateTfCommand)

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

// Only prompt for an input if the CLI is interactive and validation is true
func SurveyQuestionWithValidation(validation bool, p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	if validation {
		return SurveyQuestionInteractiveOnly(p, response, opts...)
	}
	return nil
}

type SurveyQuestionWithValidationArgs struct {
	Prompt     survey.Prompt
	Validation bool
	Response   interface{}
	Opts       []survey.AskOpt
}

// Prompt for many values at once
//
// validation: Can be used to skip the entire set of questions
func SurveyMultipleQuestionWithValidation(validation bool, questions []SurveyQuestionWithValidationArgs) error {
	if validation {
		for _, qs := range questions {
			if err := SurveyQuestionWithValidation(qs.Validation, qs.Prompt, qs.Response, qs.Opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

// Only prompt for an inputs if the CLI is interactive
func WrappedAsk(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	if !cli.nonInteractive {
		err := survey.Ask(qs, response, append(opts, survey.WithIcons(promptIconsFunc))...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Only prompt for an inputs if the CLI is interactive and validation is true
func WrappedAskValidation(validation bool, qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	if validation {
		return WrappedAsk(qs, response, opts...)
	}

	return nil
}
