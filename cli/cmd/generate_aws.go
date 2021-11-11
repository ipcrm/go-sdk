package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lacework/go-sdk/generate"
)

func promptAwsCtQuestions(config *generate.GenerateAwsTfConfigurationArgs) error {
	ctQuestions := []*survey.Question{
		{
			Name:   "useConsolidatedCloudtrail",
			Prompt: &survey.Confirm{Message: "Use consolidated Cloudtrail?"},
		},
		{
			// TODO add validator
			Name:     "awsRegion",
			Prompt:   &survey.Input{Message: "Specify the AWS region Cloudtrail, SNS, and S3 resources should use:"},
			Validate: survey.Required,
		},
		{
			Name:   "existingBucketArn",
			Prompt: &survey.Input{Message: "(Optional) Specify an existing bucket ARN used for Cloudtrail logs:"},
		},
		{
			Name:   "useExistingIamRole",
			Prompt: &survey.Confirm{Message: "(Optional) Use an existing IAM Role?"},
		},
	}

	ctAnswers := struct {
		AwsRegion                 string `survey:"awsRegion"`
		ExistingBucketArn         string `survey:"existingBucketArn"`
		ExistingSnsTopicName      string `survey:"existingSnsTopicName"`
		UseConsolidatedCloudtrail bool   `survey:"useConsolidatedCloudtrail"`
		UseExistingIamRole        bool   `survey:"useExistingIamRole"`
	}{}

	if config.ConfigureCloudtrail && !cli.nonInteractive {
		err := survey.Ask(ctQuestions, &ctAnswers,
			survey.WithIcons(promptIconsFunc),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func promptAwsExistingIamQuestions(config *generate.GenerateAwsTfConfigurationArgs) error {
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
	if config.UseExistingIamRole && !cli.nonInteractive {
		err := survey.Ask(ctExistingIamQuestions, &ctExistingIamAnswers,
			survey.WithIcons(promptIconsFunc),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func promptAwsAdditionalAccountQuestions(config *generate.GenerateAwsTfConfigurationArgs) error {
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
	askAgain := true

	// Determine the profile for the main account
	err := survey.AskOne(
		// TODO Make this prompt better
		&survey.Input{Message: "What is the AWS profile name for the main account?"},
		&config.AwsProfile,
		survey.WithValidator(survey.Required),
	)
	if err != nil {
		return err
	}

	answers := accountAnswers{}
	for askAgain {
		err := survey.Ask(accountQuestions, &answers)
		if err != nil {
			return err
		}
		accountDetails[answers.AccountProfileName] = answers.AccountProfileRegion

		err = survey.AskOne(
			&survey.Confirm{Message: "Add another AWS account?"},
			&askAgain,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func promptAwsGenerate(
	config *generate.GenerateAwsTfConfigurationArgs) (string, error) {

	if !config.ConfigureConfigCli {
		err := WrappedAskOne(&survey.Confirm{Message: "Enable Config Integration?"}, &config.ConfigureConfig)
		if err != nil {
			return "", err
		}
	} else {
		config.ConfigureConfig = true
	}

	if !config.ConfigureCloudtrailCli {
		err := WrappedAskOne(&survey.Confirm{Message: "Enable Cloudtrail Integration?"}, &config.ConfigureCloudtrail)
		if err != nil {
			return "", err
		}
	} else {
		config.ConfigureCloudtrail = true
	}

	// Set CT Specific values
	err := promptAwsCtQuestions(config)
	if err != nil {
		return "", err
	}

	// Set Existing IAM Role values
	err = promptAwsExistingIamQuestions(config)
	if err != nil {
		return "", err
	}

	// If a new bucket is to be created; should the force destroy bit be set?
	if config.ExistingBucketArn != "" {
		err := survey.AskOne(
			&survey.Confirm{Message: "Should the new S3 bucket have force destroy enabled?"},
			&config.ForceDestroyS3Bucket)
		if err != nil {
			return "", err
		}
	}

	// Let's collect up the other AWS accounts they would like to support
	if config.UseConsolidatedCloudtrail { // TODO This isn't the only time there might be sub-accounts
		err := survey.AskOne(
			&survey.Confirm{Message: "Are there additional AWS accounts to intergrate for Configuration?"},
			&config.ConfigureMoreAccounts)
		if err != nil {
			return "", err
		}

		if config.ConfigureMoreAccounts && !cli.nonInteractive {
			promptAwsAdditionalAccountQuestions(config)
		}
	}

	// Generate TF Code
	cli.StartProgress("Generating Terraform Code...")
	hcl := generate.NewAwsTFConfiguration(config)

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
