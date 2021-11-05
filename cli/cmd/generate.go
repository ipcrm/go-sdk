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

	err := survey.Ask(questions, &answers,
		survey.WithIcons(promptIconsFunc),
	)

	cli.StartProgress(" Generating Terraform Code...")
	hcl := generate.NewAwsTFConfiguration(&generate.GenerateAwsTfConfigurationArgs{
		ConfigureCloudtrail: answers.ConfigureCloudtrail,
		ConfigureConfig:     answers.ConfigureConfig,
	})

	cli.StopProgress()
	return hcl, err
}
