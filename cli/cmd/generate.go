package cmd

import (
	"fmt"

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
	}

	// aws command is used to generate TF code for aws
	generateAwsTfCommand = &cobra.Command{
		Use:   "aws",
		Short: "generate code for aws environment",
		Long:  "Genereate Terraform code for deploying into a new AWS enviornment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			location, err := promptAwsGenerate(generate.GenerateAwsCommandState)
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

	// add sub-commands to the iac-generate command
	generateTfCommand.AddCommand(generateAwsTfCommand)
}

func WrappedAskOne(p survey.Prompt, response interface{}) error {
	if !cli.nonInteractive {
		err := survey.AskOne(p, response)
		if err != nil {
			return err
		}
	}
	return nil
}
