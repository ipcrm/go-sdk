package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/lacework/go-sdk/generate"
	"github.com/pkg/errors"
)

func promptAwsCtQuestions(config *generate.GenerateAwsTfConfiguration) error {
	// Only ask these questions if configure cloudtrail is true
	if err := SurveyMultipleQuestionWithValidation([]SurveyQuestionWithValidationArgs{
		{
			Prompt:   &survey.Confirm{Message: "Use consolidated Cloudtrail?", Default: config.UseConsolidatedCloudtrail},
			Response: &config.UseConsolidatedCloudtrail,
		},
		{
			Prompt:   &survey.Confirm{Message: "Use existing cloudtrail?", Default: config.UseExistingCloudtrail},
			Response: &config.UseExistingCloudtrail,
		},
		{
			Prompt: &survey.Input{
				Message: "Specify an existing bucket ARN used for Cloudtrail logs:",
				Default: config.ExistingBucketArn,
			},
			Checks:   []*bool{&config.UseExistingCloudtrail},
			Required: true,
			Response: &config.ExistingBucketArn,
		},
	}, config.ConfigureCloudtrail); err != nil {
		return err
	}

	// If a new bucket is to be created; should the force destroy bit be set?
	newBucket := config.ExistingBucketArn == ""
	if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
		Prompt: &survey.Confirm{
			Message: "Should the new S3 bucket have force destroy enabled?",
			Default: config.ForceDestroyS3Bucket},
		Response: &config.ForceDestroyS3Bucket,
		Checks:   []*bool{&config.ConfigureCloudtrail, &newBucket}}); err != nil {
		return err
	}

	// Validate that at least region was set
	if config.ConfigureCloudtrail && config.AwsRegion == "" {
		return errors.New("AWS Region must be set when configuring Cloudtrail!")
	}

	// Validate if using an existing cloudtrail the bucket was provided
	if config.UseExistingCloudtrail && config.ExistingBucketArn == "" {
		return errors.New("Must supply bucket ARN when using an existing cloudtrail!")
	}

	return nil
}

func promptAwsExistingIamQuestions(config *generate.GenerateAwsTfConfiguration) error {
	if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
		Checks:   []*bool{&config.ConfigureCloudtrail},
		Prompt:   &survey.Confirm{Message: "(Optional) Use an existing IAM Role?"},
		Response: &config.UseExistingIamRole,
	}); err != nil {
		return err
	}

	if err := SurveyMultipleQuestionWithValidation([]SurveyQuestionWithValidationArgs{
		{
			Prompt: &survey.Input{
				Message: "Specify an existing IAM role name for Cloudtrail access",
				Default: config.ExistingIamRoleName},
			Response: &config.ExistingIamRoleName,
			Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		},
		{
			Prompt: &survey.Input{
				Message: "Specify an existing IAM role ARN for Cloudtrail access",
				Default: config.ExistingIamRoleArn,
			},
			Response: &config.ExistingIamRoleArn,
			Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		},
		{
			Prompt: &survey.Input{
				Message: "Specify the external ID to be used with the existing IAM role",
				Default: config.ExistingIamRoleExternalId,
			},
			Response: &config.ExistingIamRoleExternalId,
			Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)}, // TODO @ipcrm add validator
		}}, config.UseExistingIamRole); err != nil {
		return err
	}

	// Validate required values got set one way or another
	// If this was run non-interactive and parts of the data are missing, error out
	if config.UseExistingIamRole {
		if config.ExistingIamRoleArn == "" ||
			config.ExistingIamRoleName == "" ||
			config.ExistingIamRoleExternalId == "" {
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
	if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
		Prompt: &survey.Input{
			Message: "What is the AWS profile name for the main account?", // TODO @ipcrm Make this prompt better
			Default: config.AwsProfile,
		},
		Response: &config.AwsProfile,
		Required: true}); err != nil {
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

func promptAwsAdditionalAccounts(config *generate.GenerateAwsTfConfiguration) error {
	// Let's collect up the other AWS accounts they would like to support
	// TODO @ipcrm This isn't the only time there might be sub-accounts
	if err := SurveyQuestionInteractiveOnly(SurveyQuestionWithValidationArgs{
		Checks:   []*bool{&config.UseConsolidatedCloudtrail},
		Prompt:   &survey.Confirm{Message: "Are there additional AWS accounts to intergrate for Configuration?", Default: false},
		Response: &config.ConfigureMoreAccounts}); err != nil {
		return err
	}

	if config.ConfigureMoreAccounts && !cli.nonInteractive {
		promptAwsAdditionalAccountQuestions(config)
	}

	return nil
}

func setValsFromCliInput(config *generate.GenerateAwsTfConfiguration) {
	// Determine if configuring 'config' integration was passed
	if config.ConfigureConfigCli {
		config.ConfigureConfig = true
	}

	// Determine if configuring 'cloudtrail' integration was passed
	if config.ConfigureCloudtrailCli {
		config.ConfigureCloudtrail = true
	}

	// If config.ConsolidatedCtCli was supplied, enable ConsolidatedCt // TODO remove
	if config.ConsolidatedCtCli {
		config.UseConsolidatedCloudtrail = true
	}

	// If a bucket arn was supplied, we are using an existing ct
	if config.ExistingBucketArn != "" {
		config.UseExistingCloudtrail = true
	}

	// If any of these were set in the command line args we need to set useexistingiamrole to true and prompt for what we
	// are missing
	if config.ExistingIamRoleArn != "" ||
		config.ExistingIamRoleName != "" ||
		config.ExistingIamRoleExternalId != "" {
		config.UseExistingIamRole = true
	}

}

func promptAwsGenerate(config *generate.GenerateAwsTfConfiguration) error {
	// Set vals that were passed in, where required
	setValsFromCliInput(config)

	// This are the core questions that should be asked.  Region required for provider block
	if err := SurveyMultipleQuestionWithValidation(
		[]SurveyQuestionWithValidationArgs{
			{
				Prompt:   &survey.Confirm{Message: "Enable Config Integration?", Default: config.ConfigureConfig},
				Response: &config.ConfigureConfig,
			},
			{
				Prompt:   &survey.Confirm{Message: "Enable Cloudtrail Integration?", Default: config.ConfigureCloudtrail},
				Response: &config.ConfigureCloudtrail,
			},
			{
				Checks: []*bool{&config.ConfigureConfig, &config.ConfigureCloudtrail},
				Prompt: &survey.Input{
					Message: "Specify the AWS region Cloudtrail, SNS, and S3 resources should use:",
					Default: config.AwsRegion,
				},
				Response: &config.AwsRegion,
				Opts:     []survey.AskOpt{survey.WithValidator(survey.Required)},
			},
		}); err != nil {
		return err
	}

	// Validate one of config or cloudtrail was enabled; otherwise error out
	if !config.ConfigureCloudtrail && !config.ConfigureConfig {
		return errors.New("Must enable cloudtrail or config!")
	}

	// Set CT Specific values
	if err := promptAwsCtQuestions(config); err != nil {
		return err
	}

	// Set Existing IAM Role values
	if err := promptAwsExistingIamQuestions(config); err != nil {
		return nil
	}

	// Setup additional accounts, as required
	if err := promptAwsAdditionalAccounts(config); err != nil {
		return err
	}

	return nil
}
