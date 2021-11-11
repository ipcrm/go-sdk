package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/lacework/go-sdk/generate"
	"github.com/pkg/errors"
)

func promptAwsCtQuestions(config *generate.GenerateAwsTfConfiguration) error {
	// Set vals from CLI
	if config.ConsolidatedCtCli {
		config.UseConsolidatedCloudtrail = true
	}

	// Evaulate the rest of the params
	if config.ConfigureCloudtrail {
		if err := SurveyQuestionWithValidation(
			!config.UseConsolidatedCloudtrail,
			&survey.Confirm{Message: "Use consolidated Cloudtrail?"},
			&generate.GenerateAwsCommandState.UseConsolidatedCloudtrail); err != nil {
			return err
		}

		if err := SurveyQuestionWithValidation(
			config.AwsRegion == "",
			&survey.Input{Message: "Specify the AWS region Cloudtrail, SNS, and S3 resources should use:"},
			&generate.GenerateAwsCommandState.AwsProfile,
			survey.WithValidator(survey.Required)); err != nil { // TODO @ipcrm add validator for region
			return err
		}

		if err := SurveyQuestionWithValidation(
			config.ExistingBucketArn == "",
			&survey.Input{Message: "(Optional) Specify an existing bucket ARN used for Cloudtrail logs:"},
			&generate.GenerateAwsCommandState.ExistingBucketArn); err != nil { //TODO @ipcrm add validator
			return err
		}
	}

	return nil
}

func promptAwsExistingIamQuestions(config *generate.GenerateAwsTfConfiguration) error {
	// If any of these were set in the command line args we need to set useexistingiamrole to true and prompt for what we
	// are missing
	if generate.GenerateAwsCommandState.ExistingIamRoleArn != "" ||
		generate.GenerateAwsCommandState.ExistingIamRoleName != "" ||
		generate.GenerateAwsCommandState.ExistingIamRoleExternalId != "" {
		generate.GenerateAwsCommandState.UseExistingIamRole = true
	}

	if err := SurveyQuestionWithValidation(
		!generate.GenerateAwsCommandState.UseExistingIamRole,
		&survey.Confirm{Message: "(Optional) Use an existing IAM Role?"},
		&generate.GenerateAwsCommandState.UseExistingIamRole); err != nil {
		return err
	}

	if err := SurveyMultipleQuestionWithValidation([]SurveyQuestionWithValidationArgs{
		{
			Prompt:     &survey.Input{Message: "Specify an existing IAM role name for Cloudtrail access"},
			Response:   &generate.GenerateAwsCommandState.ExistingIamRoleName,
			Validation: generate.GenerateAwsCommandState.ExistingIamRoleName == "",
			Opts:       []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		},
		{
			Prompt:     &survey.Input{Message: "Specify an existing IAM role ARN for Cloudtrail access"},
			Response:   &generate.GenerateAwsCommandState.ExistingIamRoleArn,
			Validation: generate.GenerateAwsCommandState.ExistingIamRoleArn == "",
			Opts:       []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		},
		{
			Prompt:     &survey.Input{Message: "Specify the external ID to be used with the existing IAM role"},
			Response:   &generate.GenerateAwsCommandState.ExistingIamRoleExternalId,
			Validation: generate.GenerateAwsCommandState.ExistingIamRoleExternalId == "",
			Opts:       []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		}}, generate.GenerateAwsCommandState.UseExistingIamRole); err != nil {
		return err
	}

	// Validate required values got set one way or another
	// If this was run non-interactive and parts of the data are missing, error out
	if generate.GenerateAwsCommandState.UseExistingIamRole {
		if generate.GenerateAwsCommandState.ExistingIamRoleArn == "" ||
			generate.GenerateAwsCommandState.ExistingIamRoleName == "" ||
			generate.GenerateAwsCommandState.ExistingIamRoleExternalId == "" {
			return errors.New("When using an existing IAM role, the existing role ARN, Name, and External ID all must be set!")
		}
	}

	return nil
}

func promptAwsAdditionalAccountQuestions(config *generate.GenerateAwsTfConfiguration) error {
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
	if err := SurveyQuestionWithValidation(
		config.AwsProfile == "",
		&survey.Input{Message: "What is the AWS profile name for the main account?"}, // TODO @ipcrm Make this prompt better
		&config.AwsProfile,
		survey.WithValidator(survey.Required)); err != nil {
		return nil
	}

	// For each account to add, collect the aws profile and region to use
	for askAgain {
		answers := accountAnswers{}
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
	config.Profiles = accountDetails

	return nil
}

func promptAwsGenerate(config *generate.GenerateAwsTfConfiguration) error {
	// Determine if configuring 'config' integration was passed
	if config.ConfigureConfigCli {
		config.ConfigureConfig = true
	}

	// Determine if configuring 'cloudtrail' integration was passed
	if config.ConfigureCloudtrailCli {
		config.ConfigureCloudtrail = true
	}

	if err := SurveyMultipleQuestionWithValidation(
		[]SurveyQuestionWithValidationArgs{
			{
				Prompt:     &survey.Confirm{Message: "Enable Config Integration?"},
				Validation: !config.ConfigureConfigCli && !config.ConfigureConfig,
				Response:   &config.ConfigureConfig,
			},
			{
				Prompt:     &survey.Confirm{Message: "Enable Cloudtrail Integration?"},
				Validation: !config.ConfigureCloudtrailCli && !config.ConfigureCloudtrail,
				Response:   &config.ConfigureCloudtrail,
			}}); err != nil {
		return err
	}

	// Set CT Specific values
	err := promptAwsCtQuestions(config)
	if err != nil {
		return err
	}

	// Set Existing IAM Role values
	err = promptAwsExistingIamQuestions(config)
	if err != nil {
		return err
	}

	// If a new bucket is to be created; should the force destroy bit be set?
	if err := SurveyQuestionWithValidation(
		config.ExistingBucketArn != "",
		&survey.Confirm{Message: "Should the new S3 bucket have force destroy enabled?"},
		&config.ForceDestroyS3Bucket); err != nil {
		return err
	}

	// Let's collect up the other AWS accounts they would like to support
	// TODO @ipcrm This isn't the only time there might be sub-accounts
	if err := SurveyQuestionWithValidation(
		config.UseConsolidatedCloudtrail,
		&survey.Confirm{Message: "Are there additional AWS accounts to intergrate for Configuration?"},
		&config.ConfigureMoreAccounts); err != nil {
		return err
	}

	if config.ConfigureMoreAccounts && !cli.nonInteractive {
		promptAwsAdditionalAccountQuestions(config)
	}
	return nil
}
